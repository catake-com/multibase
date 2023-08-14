package main

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/catake-com/multibase/backend/handler"
	"github.com/catake-com/multibase/backend/module/grpc"
	"github.com/catake-com/multibase/backend/module/kafka"
	"github.com/catake-com/multibase/backend/module/kubernetes"
	"github.com/catake-com/multibase/backend/module/project"
	"github.com/catake-com/multibase/backend/module/thrift"
	"github.com/catake-com/multibase/backend/state"
)

type App struct {
	initialErr   error
	ctx          context.Context
	appLogger    *logrus.Logger
	stateStorage *state.Storage

	ProjectModule    *project.Module
	GRPCModule       *grpc.Module
	ThriftModule     *thrift.Module
	KafkaModule      *kafka.Module
	KubernetesModule *kubernetes.Module

	ProjectHandler *handler.ProjectHandler
}

// nolint: nonamedreturns, funlen
func NewApp(appLogger *logrus.Logger) (app *App) {
	// initialize empty structs so that the corresponding methods are still correctly bound in JS in case of error
	app = &App{
		GRPCModule:       &grpc.Module{},
		ThriftModule:     &thrift.Module{},
		KafkaModule:      &kafka.Module{},
		KubernetesModule: &kubernetes.Module{},
		ProjectHandler:   &handler.ProjectHandler{},
	}

	defer func() {
		if r := recover(); r != nil {
			// nolint: goerr113
			app.initialErr = errors.Join(
				app.initialErr,
				fmt.Errorf("recovered a panic: %s, stacktrace:\n%s", r, string(debug.Stack())),
			)
		}
	}()

	stateStorage, err := state.NewStorage(appLogger.WithField("component", "storage"))
	if err != nil {
		app.initialErr = errors.Join(app.initialErr, fmt.Errorf("failed to init a storage: %w", err))
	}

	projectModule, err := project.NewModule(stateStorage)
	if err != nil {
		app.initialErr = errors.Join(app.initialErr, fmt.Errorf("failed to init a project module: %w", err))
	}

	grpcModule, err := grpc.NewModule(stateStorage)
	if err != nil {
		app.initialErr = errors.Join(app.initialErr, fmt.Errorf("failed to init a grpc module: %w", err))
	}

	thriftModule, err := thrift.NewModule(stateStorage)
	if err != nil {
		app.initialErr = errors.Join(app.initialErr, fmt.Errorf("failed to init a thrift module: %w", err))
	}

	kafkaModule, err := kafka.NewModule(stateStorage, appLogger)
	if err != nil {
		app.initialErr = errors.Join(app.initialErr, fmt.Errorf("failed to init a kafka module: %w", err))
	}

	kubernetesModule, err := kubernetes.NewModule(stateStorage, appLogger)
	if err != nil {
		app.initialErr = errors.Join(app.initialErr, fmt.Errorf("failed to init a kubernetes module: %w", err))
	}

	projectHandler := handler.NewProjectHandler(
		projectModule,
		grpcModule,
		thriftModule,
		kubernetesModule,
		kafkaModule,
	)

	app.appLogger = appLogger
	app.stateStorage = stateStorage
	app.ProjectModule = projectModule
	app.GRPCModule = grpcModule
	app.ThriftModule = thriftModule
	app.KafkaModule = kafkaModule
	app.KubernetesModule = kubernetesModule
	app.ProjectHandler = projectHandler

	return app
}

func (a *App) startup(ctx context.Context) {
	if a.initialErr != nil {
		_, _ = runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Multibase failed to start correctly",
			Message: fmt.Sprintf("%+v", a.initialErr),
		})

		runtime.Quit(ctx)
	}

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
