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
	ID             string                       `json:"-"`
	Forms          map[string]*StateProjectForm `json:"forms"`
	CurrentFormID  string                       `json:"currentFormID"`
	ImportPathList []string                     `json:"importPathList"`
	ProtoFileList  []string                     `json:"protoFileList"`
	Nodes          []*ProtoTreeNode             `json:"nodes"`
}

type StateProjectForm struct {
	ID               string `json:"-"`
	Address          string `json:"address"`
	SelectedMethodID string `json:"selectedMethodID"`
	Request          string `json:"request"`
	Response         string `json:"response"`
}

type Module struct {
	AppCtx         context.Context
	configFilePath string
	state          *State
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
		projects:      map[string]*Project{},
		projectsMutex: &sync.RWMutex{},
	}

	err = module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) SendRequest(projectID, formID string, address, methodID, payload string) (*State, error) {
	m.state.Projects[projectID].Forms[formID].Address = address
	m.state.Projects[projectID].Forms[formID].SelectedMethodID = methodID
	m.state.Projects[projectID].Forms[formID].Request = payload

	response, err := m.project(projectID).SendRequest(formID, address, methodID, payload)
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
	err := m.project(projectID).StopRequest(formID)
	if err != nil {
		return nil, err
	}

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) RemoveImportPath(projectID, importPath string) (*State, error) {
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

	if lo.Contains(m.state.Projects[projectID].ProtoFileList, protoFilePath) {
		return m.state, nil
	}

	importPathList := make([]string, len(m.state.Projects[projectID].ImportPathList))
	copy(importPathList, m.state.Projects[projectID].ImportPathList)

	if len(importPathList) == 0 {
		currentDir := path.Dir(protoFilePath)
		importPathList = append(importPathList, currentDir)
	}

	protoFileList := make([]string, len(m.state.Projects[projectID].ProtoFileList))
	copy(protoFileList, m.state.Projects[projectID].ProtoFileList)
	protoFileList = append(protoFileList, protoFilePath)

	nodes, err := m.project(projectID).RefreshProtoDescriptors(importPathList, protoFileList)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].ImportPathList = importPathList
	m.state.Projects[projectID].ProtoFileList = protoFileList

	return m.state, nil
}

func (m *Module) OpenImportPath(projectID string) (*State, error) {
	importPath, err := runtime.OpenDirectoryDialog(m.AppCtx, runtime.OpenDialogOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to open import path: %w", err)
	}

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
	payload, err := m.project(projectID).SelectMethod(methodID)
	if err != nil {
		return nil, err
	}

	m.state.Projects[projectID].Forms[formID].Request = payload
	m.state.Projects[projectID].Forms[formID].SelectedMethodID = methodID

	err = m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	formID := uuid.Must(uuid.NewV4()).String()

	m.state.Projects[projectID] = &StateProject{
		ID: projectID,
		Forms: map[string]*StateProjectForm{
			formID: {
				ID:      formID,
				Address: "0.0.0.0:50051",
			},
		},
		CurrentFormID: formID,
	}

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) CreateNewForm(projectID string) (*State, error) {
	formID := uuid.Must(uuid.NewV4()).String()

	m.state.Projects[projectID].Forms[formID] = &StateProjectForm{
		ID:      formID,
		Address: "0.0.0.0:50051",
	}
	m.state.Projects[projectID].CurrentFormID = formID

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

	delete(m.state.Projects[projectID].Forms, formID)
	m.state.Projects[projectID].CurrentFormID = lo.Keys(m.state.Projects[projectID].Forms)[0]

	err := m.saveState()
	if err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) State() (*State, error) {
	return m.state, nil
}

func (m *Module) readOrInitializeState() error {
	_, err := os.Stat(m.configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return m.initializeState()
		}

		return err
	}

	return m.readState()
}

func (m *Module) initializeState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func (m *Module) readState() (rerr error) {
	file, err := os.Open(m.configFilePath)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func (m *Module) saveState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func (m *Module) project(id string) *Project {
	m.projectsMutex.RLock()
	project, ok := m.projects[id]

	if ok {
		return project
	}
	m.projectsMutex.RUnlock()

	m.projectsMutex.Lock()
	project, ok = m.projects[id]

	if ok {
		return project
	}

	project = NewProject(id)
	m.projects[id] = project
	m.projectsMutex.Unlock()

	return project
}
