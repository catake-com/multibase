package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/samber/lo"
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
	headers []*StateProjectFormHeader,
) (string, error) {
	err := f.establishConnection(address)
	if err != nil {
		return "", err
	}

	responseHandler := &responseHandler{protoDescriptorSource: protoDescriptorSource}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	grpcHeaders := lo.Map(headers, func(header *StateProjectFormHeader, _ int) string {
		return fmt.Sprintf("%s: %s", header.Key, header.Value)
	})

	f.requestCancelFunc = cancelFunc

	err = grpcurl.InvokeRPC(
		ctx,
		protoDescriptorSource,
		f.connection,
		methodID,
		grpcHeaders,
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
	if f.connection == nil {
		return nil
	}

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
	protoDescriptorSource grpcurl.DescriptorSource
	response              string
}

func (h *responseHandler) OnReceiveTrailers(status *status.Status, _ metadata.MD) {
	if status.Code() == codes.OK {
		return
	}

	formatter := grpcurl.NewJSONFormatter(
		true,
		grpcurl.AnyResolverFromDescriptorSourceWithFallback(h.protoDescriptorSource),
	)

	protoDetails := status.Proto().Details
	details := make([]map[string]string, 0, len(protoDetails))

	for _, detail := range protoDetails {
		result, err := formatter(detail)
		if err != nil {
			continue
		}

		detailMap := map[string]string{}

		err = json.Unmarshal([]byte(result), &detailMap)
		if err != nil {
			continue
		}

		detailMapWithoutType := make(map[string]string, len(detailMap))

		for key, value := range detailMap {
			if key == "@type" {
				continue
			}

			detailMapWithoutType[key] = value
		}

		details = append(details, detailMapWithoutType)
	}

	responseJSON := &ResponseJSON{
		Error: &ResponseJSONError{
			Code:    status.Code().String(),
			Message: status.Message(),
			Details: details,
		},
	}

	response, err := json.Marshal(responseJSON)
	if err != nil {
		h.response = err.Error()

		return
	}

	h.response = string(response)
}

type ResponseJSON struct {
	Error *ResponseJSONError `json:"error"`
}

type ResponseJSONError struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Details []map[string]string `json:"details,omitempty"`
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
