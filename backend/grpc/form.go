package grpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Form struct {
	id                string
	address           string
	connection        *grpc.ClientConn
	requestCancelFunc context.CancelFunc
}

func NewForm(
	id string,
	address string,
) (*Form, error) {
	form := &Form{
		id:      id,
		address: address,
	}

	return form, nil
}

func (f *Form) SendRequest(
	methodID,
	address,
	payload string,
	protoDescriptorSource grpcurl.DescriptorSource,
) (string, error) {
	err := f.establishConnection(address)
	if err != nil {
		return "", err
	}

	responseHandler := &responseHandler{}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	f.requestCancelFunc = cancelFunc

	err = grpcurl.InvokeRPC(
		ctx,
		protoDescriptorSource,
		f.connection,
		methodID,
		nil,
		responseHandler,
		func(message proto.Message) error {
			err := jsonpb.UnmarshalString(payload, message)
			if err != nil {
				return fmt.Errorf("failed to unmarshal grpc request: %w", err)
			}

			return io.EOF
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to make grpc request: %w", err)
	}

	return responseHandler.response, nil
}

func (f *Form) StopCurrentRequest() {
	if f.requestCancelFunc == nil {
		return
	}

	f.requestCancelFunc()
	f.requestCancelFunc = nil
}

func (f *Form) Close() error {
	err := f.connection.Close()
	if err != nil {
		return fmt.Errorf("failed to close grpc connection: %w", err)
	}

	return nil
}

func (f *Form) establishConnection(address string) error {
	if address == f.address && f.connection != nil {
		return nil
	}

	f.address = address

	if f.connection != nil {
		err := f.connection.Close()
		if err != nil {
			return fmt.Errorf("failed to close grpc connection: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	connection, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to establish grpc connection: %w", err)
	}

	f.connection = connection

	return nil
}

type responseHandler struct {
	response string
}

func (h *responseHandler) OnReceiveTrailers(status *status.Status, _ metadata.MD) {
	if status.Code() == codes.OK {
		return
	}

	h.response = status.String()
}

func (h *responseHandler) OnResolveMethod(_ *desc.MethodDescriptor) {
}

func (h *responseHandler) OnSendHeaders(_ metadata.MD) {
}

func (h *responseHandler) OnReceiveHeaders(_ metadata.MD) {
}

func (h *responseHandler) OnReceiveResponse(message proto.Message) {
	dynamicMessage := message.(*dynamic.Message)

	responseJSON, err := dynamicMessage.MarshalJSONPB(&jsonpb.Marshaler{EmitDefaults: true, OrigName: true})
	if err != nil {
		h.response = fmt.Sprintf("cannot parse the response due to an error: %s", err)

		return
	}

	h.response = string(responseJSON)
}
