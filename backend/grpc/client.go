package grpc

import (
	"context"
	"io"
	"time"

	"github.com/davecgh/go-spew/spew"
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

type Client struct {
	address                        string
	connection                     *grpc.ClientConn
	protoDescriptorSource          grpcurl.DescriptorSource
	protoDescriptorSourceCreatedAt time.Time
}

func NewClient(
	address string,
	protoDescriptorSource grpcurl.DescriptorSource,
	protoDescriptorSourceCreatedAt time.Time,
) (*Client, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	connection, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithDisableRetry(),
		grpc.WithReturnConnectionError(),
	)
	if err != nil {
		return nil, err
	}

	client := &Client{
		address:                        address,
		connection:                     connection,
		protoDescriptorSource:          protoDescriptorSource,
		protoDescriptorSourceCreatedAt: protoDescriptorSourceCreatedAt,
	}

	return client, nil
}

func (c *Client) Address() string {
	return c.address
}

func (c *Client) ProtoDescriptorSourceCreatedAt() time.Time {
	return c.protoDescriptorSourceCreatedAt
}

func (c *Client) SendRequest(methodID, payload string) (string, error) {
	responseHandler := &responseHandler{}

	err := grpcurl.InvokeRPC(
		context.Background(),
		c.protoDescriptorSource,
		c.connection,
		methodID,
		nil,
		responseHandler,
		func(message proto.Message) error {
			err := jsonpb.UnmarshalString(payload, message)
			if err != nil {
				return err
			}

			return io.EOF
		},
	)
	if err != nil {
		return "", err
	}

	return responseHandler.response, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}

type responseHandler struct {
	response string
}

func (h *responseHandler) OnReceiveTrailers(status *status.Status, md metadata.MD) {
	spew.Dump("OnReceiveTrailers", status, md)

	if status.Code() != codes.OK {
		h.response = status.String()
	}
}

func (h *responseHandler) OnResolveMethod(md *desc.MethodDescriptor) {
	spew.Dump("OnResolveMethod", md)
}

func (h *responseHandler) OnSendHeaders(md metadata.MD) {
	spew.Dump("OnSendHeaders", md)
}

func (h *responseHandler) OnReceiveHeaders(md metadata.MD) {
	spew.Dump("OnReceiveHeaders", md)
}

func (h *responseHandler) OnReceiveResponse(msg proto.Message) {
	spew.Dump("OnReceiveResponse", msg)

	dmsg := msg.(*dynamic.Message)
	v, _ := dmsg.MarshalJSONPB(&jsonpb.Marshaler{EmitDefaults: true, OrigName: true})
	sv := string(v)

	h.response = sv
}
