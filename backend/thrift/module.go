package thrift

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

const defaultProjectSplitterWidth = 30

type Module struct {
	AppCtx context.Context

	projects      map[string]*Project
	projectsMutex sync.RWMutex
	stateStorage  *state.Storage
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		projects:     make(map[string]*Project),
		stateStorage: stateStorage,
	}

	return module, nil
}

func (m *Module) SendRequest(projectID, formID string, address, payload string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SendRequest(
		formID,
		address,
		payload,
	)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) StopRequest(projectID, formID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	project.StopRequest(formID)

	return project, nil
}

func (m *Module) OpenFilePath(projectID string) (*Project, error) {
	filePath, err := runtime.OpenFileDialog(m.AppCtx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "Thrift Files (*.thrift)", Pattern: "*.thrift;"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open thrift file: %w", err)
	}

	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	if filePath == "" {
		return project, nil
	}

	err = project.OpenFilePath(filePath)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SelectFunction(projectID, formID, functionID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SelectFunction(formID, functionID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SaveCurrentFormID(projectID, currentFormID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveCurrentFormID(currentFormID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SaveAddress(projectID, formID, address string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveAddress(formID, address)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SaveIsMultiplexed(projectID, formID string, isMultiplexed bool) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveIsMultiplexed(formID, isMultiplexed)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) AddHeader(projectID, formID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.AddHeader(formID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SaveHeaders(projectID, formID string, headers []*Header) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveHeaders(formID, headers)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) DeleteHeader(projectID, formID, headerID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.DeleteHeader(formID, headerID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SaveSplitterWidth(projectID string, splitterWidth float64) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveSplitterWidth(splitterWidth)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SaveRequestPayload(projectID, formID, requestPayload string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SaveRequestPayload(formID, requestPayload)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) CreateNewProject(projectID string) (*Project, error) {
	project, err := NewProject(projectID, m.stateStorage)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) CreateNewForm(projectID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.CreateNewForm()
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) DeleteProject(projectID string) error {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return err
	}

	m.projectsMutex.Lock()
	defer m.projectsMutex.Unlock()

	err = project.Close()
	if err != nil {
		return err
	}

	err = m.stateStorage.Delete(projectID)
	if err != nil {
		return fmt.Errorf("failed to delete a state: %w", err)
	}

	delete(m.projects, projectID)

	return nil
}

func (m *Module) RemoveForm(projectID, formID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.RemoveForm(formID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) BeautifyRequest(projectID, formID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.BeautifyRequest(formID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) Project(projectID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) fetchProject(projectID string) (*Project, error) {
	m.projectsMutex.RLock()
	project, ok := m.projects[projectID]
	m.projectsMutex.RUnlock()

	if ok {
		return project, nil
	}

	project = &Project{}

	isLoaded, err := m.stateStorage.Load(projectID, project)
	if err != nil {
		return nil, fmt.Errorf("failed to load a state: %w", err)
	}

	if !isLoaded {
		return nil, nil
	}

	project.stateStorage = m.stateStorage

	if project.FilePath != "" {
		_, err := project.GenerateServiceTreeNodes(project.FilePath)
		if err != nil {
			return nil, err
		}
	}

	m.projectsMutex.Lock()
	m.projects[projectID] = project
	m.projectsMutex.Unlock()

	return project, nil
}
