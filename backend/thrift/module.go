package thrift

import (
	"context"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

const defaultProjectSplitterWidth = 30

type State struct {
	Projects map[string]*Project `json:"projects"`
}

type Module struct {
	AppCtx context.Context

	state        *State
	stateStorage *state.Storage
	stateMutex   *sync.RWMutex
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		state: &State{
			Projects: make(map[string]*Project),
		},
		stateStorage: stateStorage,
		stateMutex:   &sync.RWMutex{},
	}

	err := module.readOrInitializeState()
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

	project := m.state.Projects[projectID]

	response, err := project.SendRequest(
		formID,
		address,
		m.state.Projects[projectID].Forms[formID].SelectedFunctionID,
		payload,
		m.state.Projects[projectID].Forms[formID].IsMultiplexed,
		m.state.Projects[projectID].Forms[formID].Headers,
	)
	if err != nil {
		m.state.Projects[projectID].Forms[formID].Response = "{}"

		return nil, err
	}

	m.state.Projects[projectID].Forms[formID].Response = response

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) StopRequest(projectID, formID string) (*State, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	project := m.state.Projects[projectID]
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

	project := m.state.Projects[projectID]

	nodes, err := project.GenerateServiceTreeNodes(filePath)
	if err != nil {
		return nil, err
	}

	if m.state.Projects[projectID].FilePath != "" {
		for _, form := range project.Forms {
			if form.ID == m.state.Projects[projectID].CurrentFormID {
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

		m.state.Projects[projectID].Forms = map[string]*Form{form.ID: form}
	}

	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].FilePath = filePath

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SelectFunction(projectID, formID, functionID string) (*State, error) {
	project := m.state.Projects[projectID]

	payload, err := project.SelectFunction(functionID)
	if err != nil {
		return nil, err
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = payload
	m.state.Projects[projectID].Forms[formID].Response = "{}"
	m.state.Projects[projectID].Forms[formID].SelectedFunctionID = functionID

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveCurrentFormID(projectID, currentFormID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentFormID = currentFormID

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, formID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Address = address

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveIsMultiplexed(projectID, formID string, isMultiplexed bool) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].IsMultiplexed = isMultiplexed

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) AddHeader(projectID, formID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Headers = append(
		m.state.Projects[projectID].Forms[formID].Headers,
		&Header{
			ID:    uuid.Must(uuid.NewV4()).String(),
			Key:   "",
			Value: "",
		},
	)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveHeaders(projectID, formID string, headers []*Header) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Headers = headers

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteHeader(projectID, formID, headerID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Headers = lo.Reject(
		m.state.Projects[projectID].Forms[formID].Headers,
		func(header *Header, _ int) bool {
			return header.ID == headerID
		},
	)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveSplitterWidth(projectID string, splitterWidth float64) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].SplitterWidth = splitterWidth

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveRequestPayload(projectID, formID, requestPayload string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = requestPayload

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	m.state.Projects[projectID] = &Project{
		ID:            projectID,
		SplitterWidth: defaultProjectSplitterWidth,
		Forms: map[string]*Form{
			formID: {
				ID:            formID,
				Address:       "0.0.0.0:9090",
				IsMultiplexed: true,
				Request:       "{}",
				Response:      "{}",
			},
		},
		CurrentFormID: formID,
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewForm(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	var headers []*Header

	address := "0.0.0.0:9090"
	if m.state.Projects[projectID].CurrentFormID != "" {
		address = m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID].Address
		headers = m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID].Headers
	}

	m.state.Projects[projectID].Forms[formID] = &Form{
		ID:       formID,
		Address:  address,
		Request:  "{}",
		Response: "{}",
		Headers:  headers,
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)
	m.state.Projects[projectID].CurrentFormID = formID

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects, projectID)

	project, ok := m.state.Projects[projectID]
	if ok {
		err := project.Close()
		if err != nil {
			return nil, err
		}

		delete(m.state.Projects, projectID)
	}

	if err := m.saveState(); err != nil {
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

	err := m.state.Projects[projectID].Forms[formID].Close()
	if err != nil {
		return nil, err
	}

	if err := m.saveState(); err != nil {
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
	isLoaded, err := m.stateStorage.Load("thrift", m.state)
	if err != nil {
		return fmt.Errorf("failed to load a state: %w", err)
	}

	if isLoaded {
		return nil
	}

	err = m.stateStorage.Save("thrift", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) saveState() error {
	err := m.stateStorage.Save("thrift", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) initializeProjects() error {
	for _, project := range m.state.Projects {
		if project.FilePath != "" {
			_, err := project.GenerateServiceTreeNodes(project.FilePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
