package main

import (
	"context"

	"github.com/multibase-io/multibase/backend/grpc"
	"github.com/multibase-io/multibase/backend/project"
)

type App struct {
	ctx           context.Context
	ProjectModule *project.Module
	GRPCModule    *grpc.Module
}

func NewApp(projectModule *project.Module, grpcModule *grpc.Module) *App {
	return &App{
		ProjectModule: projectModule,
		GRPCModule:    grpcModule,
	}
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
