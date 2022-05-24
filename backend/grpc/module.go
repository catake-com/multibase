package grpc

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type OpenProtoFileResult struct {
	ProtoFilePath string `json:"protoFilePath"`
	CurrentDir    string `json:"currentDir"`
}

type Module struct {
	AppCtx        context.Context
	projects      map[int]*Project
	projectsMutex *sync.RWMutex
}

func NewModule() *Module {
	return &Module{
		projects:      map[int]*Project{},
		projectsMutex: &sync.RWMutex{},
	}
}

func (m *Module) SendRequest(projectID int, id int, address, methodID, payload string) (string, error) {
	return m.project(projectID).SendRequest(id, address, methodID, payload)
}

func (m *Module) StopRequest(projectID int, id int) error {
	return m.project(projectID).StopRequest(id)
}

func (m *Module) OpenProtoFile() (*OpenProtoFileResult, error) {
	protoFilePath, err := runtime.OpenFileDialog(m.AppCtx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "Proto Files (*.proto)", Pattern: "*.proto;"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open proto file: %w", err)
	}

	return &OpenProtoFileResult{
		ProtoFilePath: protoFilePath,
		CurrentDir:    path.Dir(protoFilePath),
	}, nil
}

func (m *Module) OpenImportPath() (string, error) {
	path, err := runtime.OpenDirectoryDialog(m.AppCtx, runtime.OpenDialogOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to open import path: %w", err)
	}

	return path, nil
}

func (m *Module) RefreshProtoDescriptors(
	projectID int,
	importPathList,
	protoFileList []string,
) ([]*ProtoTreeNode, error) {
	nodes, err := m.project(projectID).RefreshProtoDescriptors(importPathList, protoFileList)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (m *Module) SelectMethod(projectID int, methodID string) (string, error) {
	return m.project(projectID).SelectMethod(methodID)
}

func (m *Module) project(id int) *Project {
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
