package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/multibase-io/multibase/backend/grpc"
	"github.com/multibase-io/multibase/backend/kafka"
	"github.com/multibase-io/multibase/backend/pkg/state"
	"github.com/multibase-io/multibase/backend/project"
	"github.com/multibase-io/multibase/backend/thrift"
)

type App struct {
	ctx           context.Context
	appLogger     *logrus.Logger
	stateStorage  *state.Storage
	ProjectModule *project.Module
	GRPCModule    *grpc.Module
	ThriftModule  *thrift.Module
	KafkaModule   *kafka.Module
}

func NewApp(appLogger *logrus.Logger) (*App, error) {
	stateStorage, err := state.NewStorage(appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to init a storage: %w", err)
	}

	projectModule, err := project.NewModule(stateStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to init a project module: %w", err)
	}

	grpcModule, err := grpc.NewModule(stateStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to init a grpc module: %w", err)
	}

	thriftModule, err := thrift.NewModule(stateStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to init a thrift module: %w", err)
	}

	kafkaModule, err := kafka.NewModule(stateStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to init a kafka module: %w", err)
	}

	return &App{
		stateStorage:  stateStorage,
		appLogger:     appLogger,
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
	if err := a.stateStorage.Close(); err != nil {
		a.appLogger.Println(err)
	}

	a.KafkaModule.Close()

	return false
}

func (a *App) shutdown(_ context.Context) {
}
