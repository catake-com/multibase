package grpc

import (
	"errors"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
)

type Handler struct {
	grpcClient                     *Client
	protoTree                      *ProtoTree
	protoDescriptorSource          grpcurl.DescriptorSource
	protoDescriptorSourceCreatedAt time.Time
}

func NewHandler() *Handler {

	return &Handler{}
}

func (h *Handler) SendRequest(address, methodID, payload string) (string, error) {
	err := h.initGRPCConnection(address)
	if err != nil {
		return "", err
	}

	return h.grpcClient.SendRequest(methodID, payload)
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

func (h *Handler) initGRPCConnection(address string) error {
	if address == "" {
		return errors.New("specify address")
	}

	if h.grpcClient != nil {
		if address == h.grpcClient.Address() {
			return nil
		}

		if h.protoDescriptorSourceCreatedAt.Equal(h.grpcClient.ProtoDescriptorSourceCreatedAt()) {
			return nil
		}

		err := h.grpcClient.Close()
		if err != nil {
			return err
		}
	}

	grpcClient, err := NewClient(address, h.protoDescriptorSource, h.protoDescriptorSourceCreatedAt)
	if err != nil {
		return err
	}

	h.grpcClient = grpcClient

	return nil
}
