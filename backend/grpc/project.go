package grpc

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

var (
	errNoGRPCClient   = errors.New("no grpc client exist")
	errSpecifyAddress = errors.New("specify address")
)

type Project struct {
	id                             string
	grpcClients                    map[string]*Client
	grpcClientsMutex               *sync.RWMutex
	protoTree                      *ProtoTree
	protoDescriptorSource          grpcurl.DescriptorSource
	protoDescriptorSourceCreatedAt time.Time
}

func NewProject(id string) *Project {
	return &Project{
		id:               id,
		grpcClients:      make(map[string]*Client),
		grpcClientsMutex: &sync.RWMutex{},
	}
}

func (p *Project) SendRequest(id, address, methodID, payload string) (string, error) {
	err := p.initGRPCConnection(id, address)
	if err != nil {
		return "", err
	}

	p.grpcClientsMutex.RLock()
	defer p.grpcClientsMutex.RUnlock()
	grpcClient := p.grpcClients[id]

	return grpcClient.SendRequest(methodID, payload)
}

func (p *Project) StopRequest(id string) error {
	p.grpcClientsMutex.RLock()
	defer p.grpcClientsMutex.RUnlock()
	grpcClient := p.grpcClients[id]

	if grpcClient == nil {
		return errNoGRPCClient
	}

	grpcClient.StopCurrentRequest()

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
	p.protoDescriptorSourceCreatedAt = time.Now().UTC()

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
	for _, client := range p.grpcClients {
		err := client.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) initGRPCConnection(id, address string) error {
	if address == "" {
		return errSpecifyAddress
	}

	p.grpcClientsMutex.Lock()
	defer p.grpcClientsMutex.Unlock()

	isConnectionActive, err := p.isExistingConnectionActive(id, address)
	if err != nil {
		return err
	}

	if isConnectionActive {
		return nil
	}

	grpcClient, err := NewClient(id, address, p.protoDescriptorSource, p.protoDescriptorSourceCreatedAt)
	if err != nil {
		return err
	}

	p.grpcClients[id] = grpcClient

	return nil
}

func (p *Project) isExistingConnectionActive(id, address string) (bool, error) {
	grpcClient := p.grpcClients[id]

	if grpcClient == nil {
		return false, nil
	}

	if address == grpcClient.Address() &&
		p.protoDescriptorSourceCreatedAt.Equal(grpcClient.ProtoDescriptorSourceCreatedAt()) {
		return true, nil
	}

	err := grpcClient.Close()
	if err != nil {
		return false, fmt.Errorf("failed to close grpc client: %w", err)
	}

	return false, nil
}
