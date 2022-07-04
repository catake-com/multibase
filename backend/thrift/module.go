package thrift

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/adrg/xdg"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/multierr"
)

type State struct {
	Projects map[string]*StateProject `json:"projects"`
}

type StateProject struct {
	ID            string                       `json:"id"`
	SplitterWidth float64                      `json:"splitterWidth"`
	Forms         map[string]*StateProjectForm `json:"forms"`
	FormIDs       []string                     `json:"formIDs"`
	CurrentFormID string                       `json:"currentFormID"`
	FilePath      string                       `json:"filePath"`
	Nodes         []*ServiceTreeNode           `json:"nodes"`
}

type StateProjectForm struct {
	ID                 string `json:"id"`
	Address            string `json:"address"`
	SelectedFunctionID string `json:"selectedFunctionID"`
	Request            string `json:"request"`
	Response           string `json:"response"`
}

type Module struct {
	AppCtx         context.Context
	configFilePath string
	state          *State
	stateMutex     *sync.RWMutex
	projects       map[string]*Project
}

func NewModule() (*Module, error) {
	configFilePath, err := xdg.ConfigFile("multibase/thrift.json")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve thrift config path: %w", err)
	}

	module := &Module{
		configFilePath: configFilePath,
		state: &State{
			Projects: make(map[string]*StateProject),
		},
		stateMutex: &sync.RWMutex{},
		projects:   map[string]*Project{},
	}

	err = module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	err = module.initializeProjects()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) SendRequest(projectID, formID string, address, payload string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Address = address
	m.state.Projects[projectID].Forms[formID].Request = payload

	project := m.projects[projectID]

	response, err := project.SendRequest(
		formID,
		address,
		m.state.Projects[projectID].Forms[formID].SelectedFunctionID,
		payload,
	)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].Forms[formID].Response = response

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) StopRequest(projectID, formID string) (*State, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	project := m.projects[projectID]
	project.StopRequest(formID)

	return m.state, nil
}

func (m *Module) OpenFilePath(projectID string) (*State, error) {
	filePath, err := runtime.OpenFileDialog(m.AppCtx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "Thrift Files (*.thrift)", Pattern: "*.thrift;"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open thrift file: %w", err)
	}

	if filePath == "" {
		return m.state, nil
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	project := m.projects[projectID]

	nodes, err := project.GenerateServiceTreeNodes(filePath)
	if err != nil {
		return nil, err
	}

	if m.state.Projects[projectID].FilePath != "" {
		for _, form := range project.forms {
			if form.id == m.state.Projects[projectID].CurrentFormID {
				continue
			}

			err := form.Close()
			if err != nil {
				return nil, err
			}
		}

		form := m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID]
		form.SelectedFunctionID = ""
		form.Request = "{}"
		form.Response = "{}"

		m.state.Projects[projectID].Forms = map[string]*StateProjectForm{form.ID: form}
	}

	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].FilePath = filePath

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SelectFunction(projectID, formID, functionID string) (*State, error) {
	project := m.projects[projectID]

	payload, err := project.SelectFunction(functionID)
	if err != nil {
		return nil, err
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = payload
	m.state.Projects[projectID].Forms[formID].Response = "{}"
	m.state.Projects[projectID].Forms[formID].SelectedFunctionID = functionID

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveCurrentFormID(projectID, currentFormID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentFormID = currentFormID

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, formID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Address = address

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveSplitterWidth(projectID string, splitterWidth float64) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].SplitterWidth = splitterWidth

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveRequestPayload(projectID, formID, requestPayload string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = requestPayload

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	m.state.Projects[projectID] = &StateProject{
		ID:            projectID,
		SplitterWidth: 30,
		Forms: map[string]*StateProjectForm{
			formID: {
				ID:       formID,
				Address:  "0.0.0.0:9090",
				Request:  "{}",
				Response: "{}",
			},
		},
		CurrentFormID: formID,
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)

	project := NewProject(projectID)

	err := project.InitializeForm(formID)
	if err != nil {
		return nil, err
	}

	m.projects[projectID] = project

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewForm(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	address := "0.0.0.0:9090"
	if m.state.Projects[projectID].CurrentFormID != "" {
		address = m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID].Address
	}

	m.state.Projects[projectID].Forms[formID] = &StateProjectForm{
		ID:       formID,
		Address:  address,
		Request:  "{}",
		Response: "{}",
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)
	m.state.Projects[projectID].CurrentFormID = formID

	err := m.projects[projectID].InitializeForm(formID)
	if err != nil {
		return nil, err
	}

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects, projectID)

	project, ok := m.projects[projectID]
	if ok {
		err := project.Close()
		if err != nil {
			return nil, err
		}

		delete(m.projects, projectID)
	}

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) RemoveForm(projectID, formID string) (*State, error) {
	if len(m.state.Projects[projectID].Forms) <= 1 {
		return m.state, nil
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects[projectID].Forms, formID)
	m.state.Projects[projectID].FormIDs = lo.Reject(
		m.state.Projects[projectID].FormIDs,
		func(fID string, _ int) bool {
			return formID == fID
		},
	)

	if m.state.Projects[projectID].CurrentFormID == formID {
		m.state.Projects[projectID].CurrentFormID = lo.Keys(m.state.Projects[projectID].Forms)[0]
	}

	err := m.projects[projectID].forms[formID].Close()
	if err != nil {
		return nil, err
	}

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) State() (*State, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	return m.state, nil
}

func (m *Module) readOrInitializeState() error {
	_, err := os.Stat(m.configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to describe a thrift config file: %w", err)
		}

		return m.initializeState()
	}

	return m.readState()
}

func (m *Module) initializeState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create a thrift config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	encoder := json.NewEncoder(file)

	err = encoder.Encode(m.state)
	if err != nil {
		return fmt.Errorf("failed to encode a thrift state: %w", err)
	}

	return nil
}

func (m *Module) readState() (rerr error) {
	file, err := os.Open(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open a thrift config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&m.state)
	if err != nil {
		return fmt.Errorf("failed to decode a thrift state: %w", err)
	}

	return nil
}

func (m *Module) saveState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create/truncate a thrift config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	encoder := json.NewEncoder(file)

	err = encoder.Encode(m.state)
	if err != nil {
		return fmt.Errorf("failed to encode a thrift state: %w", err)
	}

	return nil
}

func (m *Module) initializeProjects() error {
	for _, stateProject := range m.state.Projects {
		project := NewProject(stateProject.ID)

		if stateProject.FilePath != "" {
			_, err := project.GenerateServiceTreeNodes(stateProject.FilePath)
			if err != nil {
				return err
			}
		}

		for _, form := range stateProject.Forms {
			err := project.InitializeForm(form.ID)
			if err != nil {
				return err
			}
		}

		m.projects[stateProject.ID] = project
	}

	return nil
}
