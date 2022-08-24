package main

import (
	"context"
	"fmt"
	"log"

	"github.com/multibase-io/multibase/backend/grpc"
	"github.com/multibase-io/multibase/backend/kafka"
	"github.com/multibase-io/multibase/backend/project"
	"github.com/multibase-io/multibase/backend/thrift"
)

type App struct {
	ctx           context.Context
	ProjectModule *project.Module
	GRPCModule    *grpc.Module
	ThriftModule  *thrift.Module
	KafkaModule   *kafka.Module
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

	kafkaModule, err := kafka.NewModule()
	if err != nil {
		return nil, fmt.Errorf("failed to init a kafka module: %w", err)
	}

	return &App{
		ProjectModule: projectModule,
		GRPCModule:    grpcModule,
		ThriftModule:  thriftModule,
		KafkaModule:   kafkaModule,
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.GRPCModule.AppCtx = ctx
	a.ThriftModule.AppCtx = ctx
	a.KafkaModule.AppCtx = ctx
}

func (a *App) domReady(_ context.Context) {
}

func (a *App) beforeClose(_ context.Context) bool {
	err := a.KafkaModule.SaveState()
	if err != nil {
		log.Println(fmt.Errorf("failed to save kafka state: %w", err))
	}

	return false
}

func (a *App) shutdown(_ context.Context) {
}
