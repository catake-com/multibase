package thrift

import (
	"fmt"

	"go.uber.org/thriftrw/compile"
)

type ServiceTree struct {
	services       []*ServiceTreeService
	functionsByIDs map[string]*ServiceTreeFunction
}

func NewServiceTree(module *compile.Module) (*ServiceTree, error) {
	serviceTree := &ServiceTree{
		functionsByIDs: make(map[string]*ServiceTreeFunction),
	}

	if module == nil {
		return serviceTree, nil
	}

	serviceTree.services = make([]*ServiceTreeService, 0, len(module.Services))

	for _, service := range module.Services {
		serviceTreeService := &ServiceTreeService{
			id:   service.Name,
			name: service.Name,
		}

		for _, function := range service.Functions {
			serviceTreeFunction := &ServiceTreeFunction{
				id:           fmt.Sprintf("%s_%s", service.Name, function.Name),
				functionName: function.Name,
				serviceName:  service.Name,
				spec:         function,
			}

			serviceTreeService.functions = append(serviceTreeService.functions, serviceTreeFunction)
			serviceTree.functionsByIDs[serviceTreeFunction.id] = serviceTreeFunction
		}

		serviceTree.services = append(serviceTree.services, serviceTreeService)
	}

	return serviceTree, nil
}

type ServiceTreeNode struct {
	ID         string             `json:"id"`
	Label      string             `json:"label"`
	Selectable bool               `json:"selectable"`
	Children   []*ServiceTreeNode `json:"children"`
}

func (t *ServiceTree) Nodes() []*ServiceTreeNode {
	nodes := make([]*ServiceTreeNode, 0, len(t.services))

	for _, service := range t.services {
		serviceNode := &ServiceTreeNode{
			ID:         service.id,
			Label:      service.name,
			Selectable: false,
		}

		nodes = append(nodes, serviceNode)

		serviceNode.Children = make([]*ServiceTreeNode, 0, len(service.functions))

		for _, function := range service.functions {
			functionNode := &ServiceTreeNode{
				ID:         function.id,
				Label:      function.functionName,
				Selectable: true,
			}

			serviceNode.Children = append(serviceNode.Children, functionNode)
		}
	}

	return nodes
}

func (t *ServiceTree) Function(id string) *ServiceTreeFunction {
	return t.functionsByIDs[id]
}

type ServiceTreeService struct {
	id        string
	name      string
	functions []*ServiceTreeFunction
}

type ServiceTreeFunction struct {
	id           string
	functionName string
	serviceName  string
	spec         *compile.FunctionSpec
}

func (f *ServiceTreeFunction) Spec() *compile.FunctionSpec {
	return f.spec
}

func (f *ServiceTreeFunction) ServiceName() string {
	return f.serviceName
}
