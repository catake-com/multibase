package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/copier"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/multierr"

	"github.com/multibase-io/multibase/backend/pkg/storage"
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
	AppCtx         context.Context
	configFilePath string
	state          *State
	stateMutex     *sync.RWMutex
	stateTimer     *time.Timer
	projects       map[string]*Project
}

func NewModule() (*Module, error) {
	configFilePath, err := xdg.ConfigFile("multibase/grpc")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve grpc config path: %w", err)
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

func (m *Module) SaveState() error {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	if m.stateTimer != nil {
		_ = m.stateTimer.Stop()
	}

	if err := m.saveStateToFile(); err != nil {
		return fmt.Errorf("failed to save state to file: %w", err)
	}

	return nil
}

func (m *Module) SendRequest(projectID, formID string, address, payload string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Address = address
	m.state.Projects[projectID].Forms[formID].Request = payload

	project := m.projects[projectID]

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

	m.saveState()

	return m.state, nil
}

func (m *Module) StopRequest(projectID, formID string) (*State, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	project := m.projects[projectID]
	project.StopRequest(formID)

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

	m.saveState()

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

	m.state.Projects[projectID].Nodes = nodes
	m.state.Projects[projectID].ImportPathList = importPathList
	m.state.Projects[projectID].ProtoFileList = protoFileList

	m.saveState()

	return m.state, nil
}

func (m *Module) DeleteAllProtoFiles(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	project := m.projects[projectID]

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

	m.saveState()

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

	m.saveState()

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

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveCurrentFormID(projectID, currentFormID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentFormID = currentFormID

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, formID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Address = address

	m.saveState()

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

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveHeaders(projectID, formID string, headers []*StateProjectFormHeader) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Headers = headers

	m.saveState()

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

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveSplitterWidth(projectID string, splitterWidth float64) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].SplitterWidth = splitterWidth

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveRequestPayload(projectID, formID, requestPayload string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Forms[formID].Request = requestPayload

	m.saveState()

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

	if err := m.saveStateToFile(); err != nil {
		return nil, fmt.Errorf("failed to save state to file: %w", err)
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

	m.saveState()

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

	m.saveState()

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

	m.saveState()

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
			return fmt.Errorf("failed to describe a grpc config file: %w", err)
		}

		return m.initializeState()
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

	data, err := json.Marshal(m.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	encryptedState, err := storage.Encrypt(storage.DefaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	_, err = file.Write(encryptedState)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
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

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read state from file: %w", err)
	}

	decryptedData, err := storage.Decrypt(storage.DefaultPassword, data)
	if err != nil {
		if errors.Is(err, storage.ErrNoData) {
			return nil
		}

		return fmt.Errorf("failed to decrypt state: %w", err)
	}

	err = json.Unmarshal(decryptedData, m.state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return nil
}

func (m *Module) saveState() {
	if m.stateTimer != nil {
		_ = m.stateTimer.Stop()
	}

	m.stateTimer = time.AfterFunc(storage.DefaultStatePersistenceDelay, func() {
		err := m.saveStateToFile()
		if err != nil {
			log.Println(fmt.Errorf("failed to save state to a file: %w", err))
		}
	})
}

func (m *Module) saveStateToFile() (rerr error) {
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

	state := &State{}

	err = copier.CopyWithOption(state, m.state, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return fmt.Errorf("failed to copy a grpc state: %w", err)
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	encryptedData, err := storage.Encrypt(storage.DefaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	_, err = file.Write(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
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
