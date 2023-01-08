package grpc

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

const defaultProjectSplitterWidth = 30

type State struct {
	Projects map[string]*StateProject `json:"projects"`
}

type StateProject struct {
	ID             string                       `json:"id"`
	SplitterWidth  float64                      `json:"splitterWidth"`
	Forms          map[string]*StateProjectForm `json:"forms"`
	FormIDs        []string                     `json:"formIDs"`
	CurrentFormID  string                       `json:"currentFormID"`
	IsReflected    bool                         `json:"isReflected"`
	ImportPathList []string                     `json:"importPathList"`
	ProtoFileList  []string                     `json:"protoFileList"`
	Nodes          []*ProtoTreeNode             `json:"nodes"`
}

type StateProjectFormHeader struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type StateProjectForm struct {
	ID               string                    `json:"id"`
	Address          string                    `json:"address"`
	Headers          []*StateProjectFormHeader `json:"headers"`
	SelectedMethodID string                    `json:"selectedMethodID"`
	Request          string                    `json:"request"`
	Response         string                    `json:"response"`
}

type Module struct {
	AppCtx       context.Context
	stateStorage *state.Storage
	state        *State
	stateMutex   *sync.RWMutex
	projects     map[string]*Project
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		stateStorage: stateStorage,
		state: &State{
			Projects: make(map[string]*StateProject),
		},
		stateMutex: &sync.RWMutex{},
		projects:   map[string]*Project{},
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

	project := m.projects[projectID]

	if m.state.Projects[projectID].IsReflected && !project.IsProtoDescriptorSourceInitialized() {
		_, err := project.ReflectProto(formID, address)
		if err != nil {
			return nil, err
		}
	}

	response, err := project.SendRequest(
		formID,
		m.state.Projects[projectID].Forms[formID].SelectedMethodID,
		address,
		payload,
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

	project := m.projects[projectID]
	project.StopRequest(formID)

	return m.state, nil
}

func (m *Module) ReflectProto(projectID, formID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	project := m.projects[projectID]

	nodes, err := project.ReflectProto(formID, address)
	if err != nil {
		return nil, err
	}

	form := m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID]
	form.SelectedMethodID = ""
	form.Request = "{}"
	form.Response = "{}"

	m.state.Projects[projectID].IsReflected = true
	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].ImportPathList = nil
	m.state.Projects[projectID].ProtoFileList = nil

	if err := m.saveState(); err != nil {
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

	if err := m.saveState(); err != nil {
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

	project := m.projects[projectID]

	nodes, err := project.RefreshProtoDescriptors(importPathList, protoFileList)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].IsReflected = false
	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].ImportPathList = importPathList
	m.state.Projects[projectID].ProtoFileList = protoFileList

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteAllProtoFiles(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	project := m.projects[projectID]

	m.state.Projects[projectID].IsReflected = false
	m.state.Projects[projectID].ProtoFileList = nil

	nodes, err := project.RefreshProtoDescriptors(
		m.state.Projects[projectID].ImportPathList,
		m.state.Projects[projectID].ProtoFileList,
	)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].Nodes = nodes

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
	form.SelectedMethodID = ""
	form.Request = "{}"
	form.Response = "{}"

	m.state.Projects[projectID].Forms = map[string]*StateProjectForm{form.ID: form}

	if err := m.saveState(); err != nil {
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

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SelectMethod(projectID, formID, methodID string) (*State, error) {
	project := m.projects[projectID]

	payload, err := project.SelectMethod(methodID)
	if err != nil {
		return nil, err
	}

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = payload
	m.state.Projects[projectID].Forms[formID].Response = "{}"
	m.state.Projects[projectID].Forms[formID].SelectedMethodID = methodID

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

func (m *Module) AddHeader(projectID, formID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Headers = append(
		m.state.Projects[projectID].Forms[formID].Headers,
		&StateProjectFormHeader{
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

func (m *Module) SaveHeaders(projectID, formID string, headers []*StateProjectFormHeader) (*State, error) {
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
		func(header *StateProjectFormHeader, _ int) bool {
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
	address := "0.0.0.0:50051"

	m.state.Projects[projectID] = &StateProject{
		ID:            projectID,
		SplitterWidth: defaultProjectSplitterWidth,
		Forms: map[string]*StateProjectForm{
			formID: {
				ID:       formID,
				Address:  address,
				Request:  "{}",
				Response: "{}",
			},
		},
		CurrentFormID: formID,
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)

	project := NewProject(projectID)

	err := project.InitializeForm(formID, address)
	if err != nil {
		return nil, err
	}

	m.projects[projectID] = project

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewForm(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	var headers []*StateProjectFormHeader

	address := "0.0.0.0:50051"
	if m.state.Projects[projectID].CurrentFormID != "" {
		address = m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID].Address
		headers = m.state.Projects[projectID].Forms[m.state.Projects[projectID].CurrentFormID].Headers
	}

	m.state.Projects[projectID].Forms[formID] = &StateProjectForm{
		ID:       formID,
		Address:  address,
		Request:  "{}",
		Response: "{}",
		Headers:  headers,
	}
	m.state.Projects[projectID].FormIDs = append(m.state.Projects[projectID].FormIDs, formID)
	m.state.Projects[projectID].CurrentFormID = formID

	err := m.projects[projectID].InitializeForm(formID, address)
	if err != nil {
		return nil, err
	}

	if err := m.saveState(); err != nil {
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

	err := m.projects[projectID].forms[formID].Close()
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
	isLoaded, err := m.stateStorage.Load("grpc", m.state)
	if err != nil {
		return fmt.Errorf("failed to load a state: %w", err)
	}

	if isLoaded {
		return nil
	}

	err = m.stateStorage.Save("grpc", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) saveState() error {
	err := m.stateStorage.Save("grpc", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) initializeProjects() error {
	for _, stateProject := range m.state.Projects {
		project := NewProject(stateProject.ID)

		if len(stateProject.ProtoFileList) > 0 {
			_, err := project.RefreshProtoDescriptors(
				stateProject.ImportPathList,
				stateProject.ProtoFileList,
			)
			if err != nil {
				return err
			}
		}

		for _, form := range stateProject.Forms {
			err := project.InitializeForm(form.ID, form.Address)
			if err != nil {
				return err
			}
		}

		m.projects[stateProject.ID] = project
	}

	return nil
}
