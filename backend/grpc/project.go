package grpc

import (
	"fmt"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

type Project struct {
	id                    string
	forms                 map[string]*Form
	protoTree             *ProtoTree
	protoDescriptorSource grpcurl.DescriptorSource
}

func NewProject(id string) *Project {
	return &Project{
		id:    id,
		forms: make(map[string]*Form),
	}
}

func (p *Project) SendRequest(id, methodID, address, payload string) (string, error) {
	form := p.forms[id]

	return form.SendRequest(methodID, address, payload, p.protoDescriptorSource)
}

func (p *Project) StopRequest(id string) {
	form := p.forms[id]

	form.StopCurrentRequest()
}

func (p *Project) InitializeForm(formID, address string) error {
	form, err := NewForm(formID, address)
	if err != nil {
		return err
	}

	p.forms[formID] = form

	return nil
}

func (p *Project) RefreshProtoDescriptors(importPathList, protoFileList []string) ([]*ProtoTreeNode, error) {
	protoDescriptorSource, err := grpcurl.DescriptorSourceFromProtoFiles(
		importPathList,
		protoFileList...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read from proto files: %w", err)
	}

	protoTree, err := NewProtoTree(protoDescriptorSource)
	if err != nil {
		return nil, err
	}

	p.protoDescriptorSource = protoDescriptorSource
	p.protoTree = protoTree

	return protoTree.Nodes(), nil
}

func (p *Project) SelectMethod(methodID string) (string, error) {
	method := p.protoTree.Method(methodID)
	methodMessage := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(method.Descriptor().GetInputType())

	methodPayloadJSON, err := methodMessage.MarshalJSONPB(&jsonpb.Marshaler{EmitDefaults: true, OrigName: true})
	if err != nil {
		return "", fmt.Errorf("failed to prepare grpc request: %w", err)
	}

	return string(methodPayloadJSON), nil
}

func (p *Project) Close() error {
	for _, form := range p.forms {
		err := form.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
