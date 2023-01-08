package grpc

import (
	"context"
	"fmt"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

type Project struct {
	ID             string           `json:"id"`
	SplitterWidth  float64          `json:"splitterWidth"`
	Forms          map[string]*Form `json:"forms"`
	FormIDs        []string         `json:"formIDs"`
	CurrentFormID  string           `json:"currentFormID"`
	IsReflected    bool             `json:"isReflected"`
	ImportPathList []string         `json:"importPathList"`
	ProtoFileList  []string         `json:"protoFileList"`
	Nodes          []*ProtoTreeNode `json:"nodes"`

	protoTree             *ProtoTree
	protoDescriptorSource grpcurl.DescriptorSource
}

func (p *Project) IsProtoDescriptorSourceInitialized() bool {
	return p.protoDescriptorSource != nil
}

func (p *Project) SendRequest(
	formID,
	methodID,
	address,
	payload string,
	headers []*Header,
) (string, error) {
	form := p.Forms[formID]

	return form.SendRequest(methodID, address, payload, p.protoDescriptorSource, headers)
}

func (p *Project) ReflectProto(formID, address string) ([]*ProtoTreeNode, error) {
	form := p.Forms[formID]

	protoDescriptorSource, err := form.ReflectProto(context.Background(), address)
	if err != nil {
		return nil, err
	}

	nodes, err := p.RefreshProtoNodes(protoDescriptorSource)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (p *Project) StopRequest(id string) {
	form := p.Forms[id]

	form.StopCurrentRequest()
}

func (p *Project) RefreshProtoDescriptors(importPathList, protoFileList []string) ([]*ProtoTreeNode, error) {
	protoDescriptorSource, err := grpcurl.DescriptorSourceFromProtoFiles(
		importPathList,
		protoFileList...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read from proto files: %w", err)
	}

	return p.RefreshProtoNodes(protoDescriptorSource)
}

func (p *Project) RefreshProtoNodes(protoDescriptorSource grpcurl.DescriptorSource) ([]*ProtoTreeNode, error) {
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

	payloadJSONStr := string(methodPayloadJSON)

	formattedJSON, err := jsbeautifier.Beautify(&payloadJSONStr, jsbeautifier.DefaultOptions())
	if err != nil {
		return "", fmt.Errorf("failed to format a method payload: %w", err)
	}

	return formattedJSON, nil
}

func (p *Project) Close() error {
	for _, form := range p.Forms {
		err := form.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
