package grpc

import (
	"errors"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

type Project struct {
	id                             int
	grpcClients                    map[int]*Client
	grpcClientsMutex               *sync.RWMutex
	protoTree                      *ProtoTree
	protoDescriptorSource          grpcurl.DescriptorSource
	protoDescriptorSourceCreatedAt time.Time
}

func NewProject(id int) *Project {
	return &Project{
		id:               id,
		grpcClients:      make(map[int]*Client),
		grpcClientsMutex: &sync.RWMutex{},
	}
}

func (p *Project) SendRequest(id int, address, methodID, payload string) (string, error) {
	err := p.initGRPCConnection(id, address)
	if err != nil {
		return "", err
	}

	p.grpcClientsMutex.RLock()
	defer p.grpcClientsMutex.RUnlock()
	grpcClient := p.grpcClients[id]

	return grpcClient.SendRequest(methodID, payload)
}

func (p *Project) StopRequest(id int) error {
	p.grpcClientsMutex.RLock()
	defer p.grpcClientsMutex.RUnlock()
	grpcClient := p.grpcClients[id]

	if grpcClient == nil {
		return errors.New("no grpc client exist")
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
		return nil, err
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
		return "", err
	}

	return string(methodPayloadJSON), nil
}

func (p *Project) initGRPCConnection(id int, address string) error {
	if address == "" {
		return errors.New("specify address")
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

func (p *Project) isExistingConnectionActive(id int, address string) (bool, error) {
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
		return false, err
	}

	return false, nil
}
