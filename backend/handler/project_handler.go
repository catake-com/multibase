package handler

import (
	"fmt"

	"github.com/catake-com/multibase/backend/module/grpc"
	"github.com/catake-com/multibase/backend/module/kafka"
	"github.com/catake-com/multibase/backend/module/kubernetes"
	"github.com/catake-com/multibase/backend/module/project"
	"github.com/catake-com/multibase/backend/module/thrift"
)

type ProjectHandler struct {
	projectModule    *project.Module
	grpcModule       *grpc.Module
	thriftModule     *thrift.Module
	kubernetesModule *kubernetes.Module
	kafkaModule      *kafka.Module
}

func NewProjectHandler(
	projectModule *project.Module,
	grpcModule *grpc.Module,
	thriftModule *thrift.Module,
	kubernetesModule *kubernetes.Module,
	kafkaModule *kafka.Module,
) *ProjectHandler {
	return &ProjectHandler{
		projectModule:    projectModule,
		grpcModule:       grpcModule,
		thriftModule:     thriftModule,
		kubernetesModule: kubernetesModule,
		kafkaModule:      kafkaModule,
	}
}

func (h *ProjectHandler) CreateGRPCProject(projectID string) (*project.Module, error) {
	_, err := h.grpcModule.CreateNewProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new grpc project: %w", err)
	}

	response, err := h.projectModule.CreateGRPCProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a grpc project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) CreateThriftProject(projectID string) (*project.Module, error) {
	_, err := h.thriftModule.CreateNewProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new thrift project: %w", err)
	}

	response, err := h.projectModule.CreateThriftProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a thrift project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) CreateKubernetesProject(projectID string) (*project.Module, error) {
	_, err := h.kubernetesModule.CreateNewProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new kubernetes project: %w", err)
	}

	response, err := h.projectModule.CreateKubernetesProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a kubernetes project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) CreateKafkaProject(projectID string) (*project.Module, error) {
	_, err := h.kafkaModule.CreateNewProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new kafka project: %w", err)
	}

	response, err := h.projectModule.CreateKafkaProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create a kafka project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) OpenProject(newProjectID, projectToOpenID string) (*project.Module, error) {
	response, err := h.projectModule.OpenProject(newProjectID, projectToOpenID)
	if err != nil {
		return nil, fmt.Errorf("failed to open a project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) DeleteProject(projectID string) (*project.Module, error) {
	response, err := h.projectModule.DeleteProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete a project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) RenameProject(projectID, name string) (*project.Module, error) {
	response, err := h.projectModule.RenameProject(projectID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to rename a project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) CreateNewProject() (*project.Module, error) {
	response, err := h.projectModule.CreateNewProject()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) CloseProject(projectID string) (*project.Module, error) {
	response, err := h.projectModule.CloseProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to close a project: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) SaveCurrentProjectID(projectID string) (*project.Module, error) {
	response, err := h.projectModule.SaveCurrentProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to save current project id: %w", err)
	}

	return response, nil
}

func (h *ProjectHandler) State() (*project.Module, error) {
	response, err := h.projectModule.State()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch a state: %w", err)
	}

	return response, nil
}
