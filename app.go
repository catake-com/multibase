package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/multibase-io/multibase/backend/grpc"
	"github.com/multibase-io/multibase/backend/kafka"
	"github.com/multibase-io/multibase/backend/kubernetes"
	"github.com/multibase-io/multibase/backend/pkg/state"
	"github.com/multibase-io/multibase/backend/project"
	"github.com/multibase-io/multibase/backend/thrift"
)

type App struct {
	ctx              context.Context
	appLogger        *logrus.Logger
	stateStorage     *state.Storage
	ProjectModule    *project.Module
	GRPCModule       *grpc.Module
	ThriftModule     *thrift.Module
	KafkaModule      *kafka.Module
	KubernetesModule *kubernetes.Module
}

func NewApp(appLogger *logrus.Logger) (*App, error) {
	stateStorage, err := state.NewStorage(appLogger.WithField("component", "storage"))
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

	kafkaModule, err := kafka.NewModule(stateStorage, appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to init a kafka module: %w", err)
	}

	kubernetesModule, err := kubernetes.NewModule(stateStorage, appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to init a kubernetes module: %w", err)
	}

	return &App{
		stateStorage:     stateStorage,
		appLogger:        appLogger,
		ProjectModule:    projectModule,
		GRPCModule:       grpcModule,
		ThriftModule:     thriftModule,
		KafkaModule:      kafkaModule,
		KubernetesModule: kubernetesModule,
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.GRPCModule.AppCtx = ctx
	a.ThriftModule.AppCtx = ctx
	a.KafkaModule.AppCtx = ctx
	a.KubernetesModule.AppCtx = ctx
}

func (a *App) domReady(_ context.Context) {
}

func (a *App) beforeClose(_ context.Context) bool {
	if err := a.stateStorage.Close(); err != nil {
		a.appLogger.Errorln(err)
	}

	if err := a.KafkaModule.Close(); err != nil {
		a.appLogger.Errorln(err)
	}

	if err := a.KubernetesModule.Close(); err != nil {
		a.appLogger.Errorln(err)
	}

	return false
}

func (a *App) shutdown(_ context.Context) {
}
