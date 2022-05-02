package grpc

import (
	"time"

	"github.com/fullstorydev/grpcurl"
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

// func (h *Handler) ProtoDescriptorSource() grpcurl.DescriptorSource {
// 	return h.protoDescriptorSource
// }
//
// func (h *Handler) AddImportPath(importPath string) error {
// 	err := h.state.AddImportPath(importPath)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = h.initProtoDescriptorSource()
// 	if err != nil {
// 		err2 := h.state.RemoveImportPath(importPath)
//
// 		return multierr.Combine(err, err2)
// 	}
//
// 	return nil
// }
//
// func (h *Handler) RemoveImportPath(importPath string) error {
// 	err := h.state.RemoveImportPath(importPath)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = h.initProtoDescriptorSource()
// 	if err != nil {
// 		err2 := h.state.AddImportPath(importPath)
//
// 		return multierr.Combine(err, err2)
// 	}
//
// 	return nil
// }
//
// func (h *Handler) AddProtoFile(protoFile string) error {
// 	err := h.state.AddProtoFile(protoFile)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = h.initProtoDescriptorSource()
// 	if err != nil {
// 		err2 := h.state.RemoveProtoFile(protoFile)
//
// 		return multierr.Combine(err, err2)
// 	}
//
// 	return nil
// }
//
// func (h *Handler) RemoveProtoFile(protoFile string) error {
// 	err := h.state.RemoveProtoFile(protoFile)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = h.initProtoDescriptorSource()
// 	if err != nil {
// 		err2 := h.state.AddProtoFile(protoFile)
//
// 		return multierr.Combine(err, err2)
// 	}
//
// 	return nil
// }
//
// func (h *Handler) RemoveAllProtoFiles() error {
// 	h.state.RemoveAllProtoFiles()
//
// 	err := h.initProtoDescriptorSource()
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (h *Handler) Address() string {
// 	return h.state.Address()
// }
//
// func (h *Handler) SetAddress(address string) error {
// 	if h.state.Address() == address {
// 		return nil
// 	}
//
// 	h.state.SetAddress(address)
//
// 	return nil
// }
//
// func (h *Handler) SendRequest(address, methodFQN, payload string) (string, error) {
// 	err := h.initGRPCConnection(address)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return h.grpcClient.SendRequest(methodFQN, payload)
// }

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

// func (h *Handler) initGRPCConnection(address string) error {
// 	if address == "" {
// 		return errors.New("specify address")
// 	}
//
// 	if h.grpcClient != nil {
// 		if h.state.Address() == h.grpcClient.Address() {
// 			return nil
// 		}
//
// 		if h.protoDescriptorSourceCreatedAt.Equal(h.grpcClient.ProtoDescriptorSourceCreatedAt()) {
// 			return nil
// 		}
//
// 		err := h.grpcClient.Close()
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	grpcClient, err := NewClient(h.state.Address(), h.protoDescriptorSource, h.protoDescriptorSourceCreatedAt)
// 	if err != nil {
// 		return err
// 	}
//
// 	h.grpcClient = grpcClient
//
// 	return nil
// }
