package grpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/catake-com/multibase/backend/state"
)

const defaultProjectSplitterWidth = 20

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

func (m *Module) ReflectProto(projectID, formID, address string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.ReflectProto(formID, address)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) RemoveImportPath(projectID, importPath string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.RemoveImportPath(importPath)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) OpenProtoFile(projectID string) (*Project, error) {
	protoFilePath, err := runtime.OpenFileDialog(m.AppCtx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "Proto Files (*.proto)", Pattern: "*.proto;"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open proto file: %w", err)
	}

	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	if protoFilePath == "" {
		return project, nil
	}

	err = project.OpenProtoFile(protoFilePath)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) DeleteAllProtoFiles(projectID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.DeleteAllProtoFiles()
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) OpenImportPath(projectID string) (*Project, error) {
	importPath, err := runtime.OpenDirectoryDialog(m.AppCtx, runtime.OpenDialogOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to open import path: %w", err)
	}

	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	if importPath == "" {
		return project, nil
	}

	err = project.OpenImportPath(importPath)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (m *Module) SelectMethod(projectID, formID, methodID string) (*Project, error) {
	project, err := m.fetchProject(projectID)
	if err != nil {
		return nil, err
	}

	err = project.SelectMethod(methodID, formID)
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

	if len(project.ProtoFileList) > 0 {
		_, err := project.RefreshProtoDescriptors(
			project.ImportPathList,
			project.ProtoFileList,
		)
		if err != nil {
			return nil, err
		}
	}

	m.projectsMutex.Lock()
	m.projects[projectID] = project
	m.projectsMutex.Unlock()

	return project, nil
}
