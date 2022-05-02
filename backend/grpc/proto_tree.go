package grpc

import (
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
)

type ProtoTree struct {
	files        []*ProtoTreeFile
	methodsByIDs map[string]*ProtoTreeMethod
	namesByIDs   map[string]string
}

func NewProtoTree(protoDescriptorSource grpcurl.DescriptorSource) (*ProtoTree, error) {
	protoTree := &ProtoTree{
		namesByIDs:   map[string]string{},
		methodsByIDs: map[string]*ProtoTreeMethod{},
	}

	if protoDescriptorSource == nil {
		return protoTree, nil
	}

	services, err := protoDescriptorSource.ListServices()
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		des, err := protoDescriptorSource.FindSymbol(service)
		if err != nil {
			return nil, err
		}

		serviceDesc := des.(*desc.ServiceDescriptor)

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

		protoTreeFile := protoTree.AddFile(serviceDesc.GetFile().GetFullyQualifiedName(), serviceDesc.GetFile().GetName())
		protoTreeFile.services = append(protoTreeFile.services, protoTreeService)
	}

	for _, file := range protoTree.files {
		protoTree.namesByIDs[file.id] = file.name

		for _, service := range file.services {
			protoTree.namesByIDs[service.id] = service.name

			for _, method := range service.methods {
				protoTree.namesByIDs[method.id] = method.name
			}
		}
	}

	return protoTree, nil
}

type ProtoTreeNode struct {
	ID       string           `json:"id"`
	Label    string           `json:"label"`
	Children []*ProtoTreeNode `json:"children"`
}

func (t *ProtoTree) Nodes() []*ProtoTreeNode {
	var nodes []*ProtoTreeNode

	for _, file := range t.files {
		fileNode := &ProtoTreeNode{
			ID:    file.id,
			Label: file.name,
		}

		nodes = append(nodes, fileNode)

		for _, service := range file.services {
			serviceNode := &ProtoTreeNode{
				ID:    service.id,
				Label: service.name,
			}

			fileNode.Children = append(fileNode.Children, serviceNode)

			for _, method := range service.methods {
				methodNode := &ProtoTreeNode{
					ID:    method.id,
					Label: method.name,
				}

				serviceNode.Children = append(serviceNode.Children, methodNode)
			}
		}
	}

	return nodes
}

func (t *ProtoTree) Name(id string) string {
	return t.namesByIDs[id]
}

func (t *ProtoTree) Method(id string) (*ProtoTreeMethod, bool) {
	method, ok := t.methodsByIDs[id]

	return method, ok
}

func (t *ProtoTree) AddFile(id, name string) *ProtoTreeFile {
	for _, file := range t.files {
		if file.id == id {
			return file
		}
	}

	file := &ProtoTreeFile{id: id, name: name}

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

func (m *ProtoTreeMethod) Name() string {
	return m.name
}

func (m *ProtoTreeMethod) Descriptor() *desc.MethodDescriptor {
	return m.descriptor
}
