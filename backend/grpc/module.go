package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
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
	ID             string                       `json:"id"`
	SplitterWidth  float64                      `json:"splitterWidth"`
	Forms          map[string]*StateProjectForm `json:"forms"`
	FormIDs        []string                     `json:"formIDs"`
	CurrentFormID  string                       `json:"currentFormID"`
	ImportPathList []string                     `json:"importPathList"`
	ProtoFileList  []string                     `json:"protoFileList"`
	Nodes          []*ProtoTreeNode             `json:"nodes"`
}

type StateProjectForm struct {
	ID               string `json:"id"`
	Address          string `json:"address"`
	SelectedMethodID string `json:"selectedMethodID"`
	Request          string `json:"request"`
	Response         string `json:"response"`
}

type Module struct {
	AppCtx         context.Context
	configFilePath string
	state          *State
	stateMutex     *sync.RWMutex
	projects       map[string]*Project
	projectsMutex  *sync.RWMutex
}

func NewModule() (*Module, error) {
	configFilePath, err := xdg.ConfigFile("multibase/grpc.json")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve grpc config path: %w", err)
	}

	module := &Module{
		configFilePath: configFilePath,
		state: &State{
			Projects: make(map[string]*StateProject),
		},
		stateMutex:    &sync.RWMutex{},
		projects:      map[string]*Project{},
		projectsMutex: &sync.RWMutex{},
	}

	err = module.readOrInitializeState()
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

	project, err := m.project(projectID)
	if err != nil {
		return nil, err
	}

	response, err := project.SendRequest(
		formID,
		address,
		m.state.Projects[projectID].Forms[formID].SelectedMethodID,
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

	project, err := m.project(projectID)
	if err != nil {
		return nil, err
	}

	err = project.StopRequest(formID)
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) RemoveImportPath(projectID, importPath string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].ImportPathList = lo.Reject(
		m.state.Projects[projectID].ImportPathList,
		func(ip string, _ int) bool {
			return ip == importPath
		},
	)

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) OpenProtoFile(projectID string) (*State, error) {
	protoFilePath, err := runtime.OpenFileDialog(m.AppCtx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "Proto Files (*.proto)", Pattern: "*.proto;"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open proto file: %w", err)
	}

	if protoFilePath == "" {
		return m.state, nil
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if lo.Contains(m.state.Projects[projectID].ProtoFileList, protoFilePath) {
		return m.state, nil
	}

	var importPathList []string
	if len(m.state.Projects[projectID].ImportPathList) > 0 {
		importPathList = m.state.Projects[projectID].ImportPathList
	} else {
		currentDir := path.Dir(protoFilePath)
		importPathList = []string{currentDir}
	}

	protoFileList := append([]string{protoFilePath}, m.state.Projects[projectID].ProtoFileList...)

	project, err := m.project(projectID)
	if err != nil {
		return nil, err
	}

	nodes, err := project.RefreshProtoDescriptors(importPathList, protoFileList)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].ImportPathList = importPathList
	m.state.Projects[projectID].ProtoFileList = protoFileList

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteAllProtoFiles(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	project, err := m.project(projectID)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].ProtoFileList = nil

	nodes, err := project.RefreshProtoDescriptors(
		m.state.Projects[projectID].ImportPathList,
		m.state.Projects[projectID].ProtoFileList,
	)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].Nodes = nodes

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) OpenImportPath(projectID string) (*State, error) {
	importPath, err := runtime.OpenDirectoryDialog(m.AppCtx, runtime.OpenDialogOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to open import path: %w", err)
	}

	if importPath == "" {
		return m.state, nil
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if lo.Contains(m.state.Projects[projectID].ImportPathList, importPath) {
		return m.state, nil
	}

	m.state.Projects[projectID].ImportPathList = append(m.state.Projects[projectID].ImportPathList, importPath)

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SelectMethod(projectID, formID, methodID string) (*State, error) {
	project, err := m.project(projectID)
	if err != nil {
		return nil, err
	}

	payload, err := project.SelectMethod(methodID)
	if err != nil {
		return nil, err
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = payload
	m.state.Projects[projectID].Forms[formID].SelectedMethodID = methodID

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
				Address:  "0.0.0.0:50051",
				Request:  "{}",
				Response: "{}",
			},
		},
		CurrentFormID: formID,
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewForm(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	address := "0.0.0.0:50051"
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

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	delete(m.state.Projects, projectID)
	m.stateMutex.Unlock()

	m.projectsMutex.Lock()
	project, ok := m.projects[projectID]

	if ok {
		err := project.Close()
		if err != nil {
			return nil, err
		}

		delete(m.projects, projectID)
	}
	m.projectsMutex.Unlock()

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

	err := m.saveState()
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
		if os.IsNotExist(err) {
			return m.initializeState()
		}

		return fmt.Errorf("failed to describe a grpc config file: %w", err)
	}

	return m.readState()
}

func (m *Module) initializeState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create a grpc config file: %w", err)
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
		return fmt.Errorf("failed to encode a grpc state: %w", err)
	}

	return nil
}

func (m *Module) readState() (rerr error) {
	file, err := os.Open(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open a grpc config file: %w", err)
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
		return fmt.Errorf("failed to decode a grpc state: %w", err)
	}

	return nil
}

func (m *Module) saveState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create/truncate a grpc config file: %w", err)
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
		return fmt.Errorf("failed to encode a grpc state: %w", err)
	}

	return nil
}

func (m *Module) project(projectID string) (*Project, error) {
	m.projectsMutex.RLock()
	project, ok := m.projects[projectID]

	if ok {
		return project, nil
	}
	m.projectsMutex.RUnlock()

	m.projectsMutex.Lock()
	project, ok = m.projects[projectID]

	if ok {
		return project, nil
	}

	project = NewProject(projectID)

	_, err := project.RefreshProtoDescriptors(
		m.state.Projects[projectID].ImportPathList,
		m.state.Projects[projectID].ProtoFileList,
	)
	if err != nil {
		return nil, err
	}

	m.projects[projectID] = project
	m.projectsMutex.Unlock()

	return project, nil
}
