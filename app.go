package main

import (
	"context"
	"fmt"

	"github.com/multibase-io/multibase/backend/grpc"
	"github.com/multibase-io/multibase/backend/project"
	"github.com/multibase-io/multibase/backend/thrift"
)

type App struct {
	ctx           context.Context
	ProjectModule *project.Module
	GRPCModule    *grpc.Module
	ThriftModule  *thrift.Module
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

	thriftModule, err := thrift.NewModule()
	if err != nil {
		return nil, fmt.Errorf("failed to init a thrift module: %w", err)
	}

	return &App{
		ProjectModule: projectModule,
		GRPCModule:    grpcModule,
		ThriftModule:  thriftModule,
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.GRPCModule.AppCtx = ctx
	a.ThriftModule.AppCtx = ctx
}

func (a *App) domReady(ctx context.Context) {
}

func (a *App) beforeClose(ctx context.Context) bool {
	return false
}

func (a *App) shutdown(ctx context.Context) {
}
