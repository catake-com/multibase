package grpc

import (
	"errors"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

type Handler struct {
	grpcClients                    map[int]*Client
	grpcClientsMutex               *sync.RWMutex
	protoTree                      *ProtoTree
	protoDescriptorSource          grpcurl.DescriptorSource
	protoDescriptorSourceCreatedAt time.Time
}

func NewHandler() *Handler {
	return &Handler{
		grpcClients:      make(map[int]*Client),
		grpcClientsMutex: &sync.RWMutex{},
	}
}

func (h *Handler) SendRequest(id int, address, methodID, payload string) (string, error) {
	err := h.initGRPCConnection(id, address)
	if err != nil {
		return "", err
	}

	h.grpcClientsMutex.RLock()
	defer h.grpcClientsMutex.RUnlock()
	grpcClient := h.grpcClients[id]

	return grpcClient.SendRequest(methodID, payload)
}

func (h *Handler) StopRequest(id int) error {
	h.grpcClientsMutex.RLock()
	defer h.grpcClientsMutex.RUnlock()
	grpcClient := h.grpcClients[id]

	if grpcClient == nil {
		errors.New("no grpc client exist")
	}

	grpcClient.StopCurrentRequest()

	return nil
}

func (h *Handler) RefreshProtoDescriptors(importPathList, protoFileList []string) ([]*ProtoTreeNode, error) {
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

	h.protoDescriptorSource = protoDescriptorSource
	h.protoTree = protoTree
	h.protoDescriptorSourceCreatedAt = time.Now().UTC()

	return protoTree.Nodes(), nil
}

func (h *Handler) SelectMethod(methodID string) (string, error) {
	method := h.protoTree.Method(methodID)
	methodMessage := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(method.Descriptor().GetInputType())

	methodPayloadJSON, err := methodMessage.MarshalJSONPB(&jsonpb.Marshaler{EmitDefaults: true, OrigName: true})
	if err != nil {
		return "", err
	}

	return string(methodPayloadJSON), nil
}

func (h *Handler) initGRPCConnection(id int, address string) error {
	if address == "" {
		return errors.New("specify address")
	}

	h.grpcClientsMutex.Lock()
	defer h.grpcClientsMutex.Unlock()

	isConnectionActive, err := h.isExistingConnectionActive(id, address)
	if err != nil {
		return err
	}

	if isConnectionActive {
		return nil
	}

	grpcClient, err := NewClient(id, address, h.protoDescriptorSource, h.protoDescriptorSourceCreatedAt)
	if err != nil {
		return err
	}

	h.grpcClients[id] = grpcClient

	return nil
}

func (h *Handler) isExistingConnectionActive(id int, address string) (bool, error) {
	grpcClient := h.grpcClients[id]

	if grpcClient == nil {
		return false, nil
	}

	if address == grpcClient.Address() &&
		h.protoDescriptorSourceCreatedAt.Equal(grpcClient.ProtoDescriptorSourceCreatedAt()) {
		return true, nil
	}

	err := grpcClient.Close()
	if err != nil {
		return false, err
	}

	return false, nil
}
