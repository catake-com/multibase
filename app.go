package main

import (
	"context"
	"fmt"

	"github.com/multibase-io/multibase/backend/grpc"
	"github.com/multibase-io/multibase/backend/project"
)

type App struct {
	ctx           context.Context
	ProjectModule *project.Module
	GRPCModule    *grpc.Module
}

func NewApp() (*App, error) {
	projectModule, err := project.NewModule()
	if err != nil {
		return nil, fmt.Errorf("failed to init a project module: %w", err)
	}

	grpcModule, err := grpc.NewModule()
	if err != nil {
		return nil, fmt.Errorf("failed to init a grpc module: %w", err)
	}

	return &App{
		ProjectModule: projectModule,
		GRPCModule:    grpcModule,
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.GRPCModule.AppCtx = ctx
}

func (a *App) domReady(ctx context.Context) {
}

func (a *App) beforeClose(ctx context.Context) bool {
	return false
}

func (a *App) shutdown(ctx context.Context) {
}
