package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/multibase-io/multibase/backend/grpc"
)

// App struct
type App struct {
	ctx        context.Context
	grpcModule *grpc.Handler
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{grpcModule: grpc.NewHandler()}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) OpenProtoFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "Proto Files (*.proto)", Pattern: "*.proto;"},
		},
	})
	if err != nil {
		return "", err
	}

	return path, nil
}

func (a *App) OpenImportPath() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		return "", err
	}

	return path, nil
}

func (a *App) RefreshProtoDescriptors(importPathList, protoFileList []string) ([]*grpc.ProtoTreeNode, error) {
	nodes, err := a.grpcModule.RefreshProtoDescriptors(importPathList, protoFileList)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
