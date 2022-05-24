package main

import (
	"context"

	"github.com/multibase-io/multibase/backend/grpc"
)

type App struct {
	ctx        context.Context
	GRPCModule *grpc.Module
}

func NewApp() *App {
	return &App{
		GRPCModule: grpc.NewModule(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.GRPCModule.AppCtx = ctx
}

func (a *App) domReady(ctx context.Context) {
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func (a *App) shutdown(ctx context.Context) {
}
