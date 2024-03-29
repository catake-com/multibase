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
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

const requestTimeout = time.Second * 5

type ResponseJSON struct {
	Error *ResponseJSONError `json:"error"`
}

type ResponseJSONError struct {
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Details []map[string]interface{} `json:"details,omitempty"`
}

type Form struct {
	ID               string    `json:"id"`
	Address          string    `json:"address"`
	Headers          []*Header `json:"headers"`
	SelectedMethodID string    `json:"selectedMethodID"`
	Request          string    `json:"request"`
	Response         string    `json:"response"`

	connection        *grpc.ClientConn
	requestCancelFunc context.CancelFunc
}

func (f *Form) SendRequest(
	methodID,
	address,
	payload string,
	protoDescriptorSource grpcurl.DescriptorSource,
	headers []*Header,
) (string, error) {
	err := f.establishConnection(context.Background(), address)
	if err != nil {
		return "", err
	}

	responseHandler := &responseHandler{protoDescriptorSource: protoDescriptorSource}

	ctx, cancelFunc := context.WithTimeout(context.Background(), requestTimeout)

	grpcHeaders := lo.Map(headers, func(header *Header, _ int) string {
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

// nolint: ireturn
func (f *Form) ReflectProto(ctx context.Context, address string) (grpcurl.DescriptorSource, error) {
	err := f.establishConnection(ctx, address)
	if err != nil {
		return nil, err
	}

	// nolint: nosnakecase
	reflectionClient := grpcreflect.NewClient(
		ctx,
		grpc_reflection_v1alpha.NewServerReflectionClient(f.connection),
	)

	protoDescriptorSource := grpcurl.DescriptorSourceFromServer(ctx, reflectionClient)

	return protoDescriptorSource, nil
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

	if err := f.connection.Close(); err != nil {
		return fmt.Errorf("failed to close grpc connection: %w", err)
	}

	return nil
}

func (f *Form) establishConnection(ctx context.Context, address string) error {
	if address == f.Address && f.connection != nil {
		return nil
	}

	f.Address = address

	if f.connection != nil {
		err := f.connection.Close()
		if err != nil {
			return fmt.Errorf("failed to close grpc connection: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
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
	details := make([]map[string]interface{}, 0, len(protoDetails))

	for _, detail := range protoDetails {
		result, err := formatter(detail)
		if err != nil {
			continue
		}

		detailMap := map[string]interface{}{}

		err = json.Unmarshal([]byte(result), &detailMap)
		if err != nil {
			continue
		}

		detailMapWithoutType := make(map[string]interface{}, len(detailMap))

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

func (h *responseHandler) OnResolveMethod(_ *desc.MethodDescriptor) {
}

func (h *responseHandler) OnSendHeaders(_ metadata.MD) {
}

func (h *responseHandler) OnReceiveHeaders(_ metadata.MD) {
}

func (h *responseHandler) OnReceiveResponse(message proto.Message) {
	dynamicMessage, ok := message.(*dynamic.Message)
	if !ok {
		h.response = fmt.Sprintf("expected dynamic message, got %T instead", message)

		return
	}

	responseJSON, err := dynamicMessage.MarshalJSONPB(&jsonpb.Marshaler{EmitDefaults: true, OrigName: true})
	if err != nil {
		h.response = fmt.Sprintf("cannot parse the response due to an error: %s", err)

		return
	}

	h.response = string(responseJSON)
}
