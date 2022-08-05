package grpc

import (
	"errors"
	"fmt"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
)

var errServiceDescriptor = errors.New("expected service descriptor")

type ProtoTree struct {
	files        []*ProtoTreeFile
	methodsByIDs map[string]*ProtoTreeMethod
}

func NewProtoTree(protoDescriptorSource grpcurl.DescriptorSource) (*ProtoTree, error) {
	protoTree := &ProtoTree{
		methodsByIDs: map[string]*ProtoTreeMethod{},
	}

	if protoDescriptorSource == nil {
		return protoTree, nil
	}

	services, err := protoDescriptorSource.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list grpc services: %w", err)
	}

	for _, service := range services {
		des, err := protoDescriptorSource.FindSymbol(service)
		if err != nil {
			return nil, fmt.Errorf("failed to find service: %w", err)
		}

		serviceDesc, ok := des.(*desc.ServiceDescriptor)
		if !ok {
			return nil, fmt.Errorf("%w, got %T instead", errServiceDescriptor, des)
		}

		protoTreeService := &ProtoTreeService{
			id:   serviceDesc.GetFullyQualifiedName(),
			name: serviceDesc.GetName(),
		}

		for _, method := range serviceDesc.GetMethods() {
			protoTreeMethod := &ProtoTreeMethod{
				id:         method.GetFullyQualifiedName(),
				name:       method.GetName(),
				descriptor: method,
			}

			protoTreeService.methods = append(protoTreeService.methods, protoTreeMethod)
			protoTree.methodsByIDs[protoTreeMethod.id] = protoTreeMethod
		}

		protoTreeFile := protoTree.AddFile(
			serviceDesc.GetFile().GetFullyQualifiedName(),
			serviceDesc.GetFile().GetName(),
		)
		protoTreeFile.services = append(protoTreeFile.services, protoTreeService)
	}

	return protoTree, nil
}

type ProtoTreeNode struct {
	ID         string           `json:"id"`
	Label      string           `json:"label"`
	Selectable bool             `json:"selectable"`
	Children   []*ProtoTreeNode `json:"children"`
}

func (t *ProtoTree) Nodes() []*ProtoTreeNode {
	nodes := make([]*ProtoTreeNode, 0, len(t.files))

	for _, file := range t.files {
		fileNode := &ProtoTreeNode{
			ID:         file.id,
			Label:      file.name,
			Selectable: false,
		}

		nodes = append(nodes, fileNode)

		for _, service := range file.services {
			serviceNode := &ProtoTreeNode{
				ID:         service.id,
				Label:      service.name,
				Selectable: false,
			}

			fileNode.Children = append(fileNode.Children, serviceNode)

			for _, method := range service.methods {
				methodNode := &ProtoTreeNode{
					ID:         method.id,
					Label:      method.name,
					Selectable: true,
				}

				serviceNode.Children = append(serviceNode.Children, methodNode)
			}
		}
	}

	return nodes
}

func (t *ProtoTree) Method(id string) *ProtoTreeMethod {
	return t.methodsByIDs[id]
}

func (t *ProtoTree) AddFile(fileID, name string) *ProtoTreeFile {
	for _, file := range t.files {
		if file.id == fileID {
			return file
		}
	}

	file := &ProtoTreeFile{id: fileID, name: name}

	t.files = append(t.files, file)

	return file
}

type ProtoTreeFile struct {
	id       string
	name     string
	services []*ProtoTreeService
}

type ProtoTreeService struct {
	id      string
	name    string
	methods []*ProtoTreeMethod
}

type ProtoTreeMethod struct {
	id         string
	name       string
	descriptor *desc.MethodDescriptor
}

func (m *ProtoTreeMethod) Descriptor() *desc.MethodDescriptor {
	return m.descriptor
}
