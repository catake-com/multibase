package thrift

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/yarpc/yab/thrift"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

const requestTimeout = 5 * time.Second

type Form struct {
	id                string
	client            *http.Client
	serviceTree       *ServiceTree
	requestCancelFunc context.CancelFunc
}

func NewForm(
	formID string,
	serviceTree *ServiceTree,
) (*Form, error) {
	client := &http.Client{
		Timeout: requestTimeout,
	}

	form := &Form{
		id:          formID,
		client:      client,
		serviceTree: serviceTree,
	}

	return form, nil
}

func (f *Form) SendRequest(functionID, address, payload string, headers []*StateProjectFormHeader) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), requestTimeout)
	f.requestCancelFunc = cancelFunc

	function := f.serviceTree.Function(functionID)

	var requestPayload map[string]interface{}

	err := yaml.Unmarshal([]byte(payload), &requestPayload)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	encodedPayload, err := thrift.RequestToBytes(
		function.Spec(),
		requestPayload,
		thrift.Options{
			UseEnvelopes:         true,
			EnvelopeMethodPrefix: fmt.Sprintf("%s:", function.ServiceName()),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to build thrift encoded payload: %w", err)
	}

	requestURL := &url.URL{Scheme: "http", Host: address}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL.String(),
		bytes.NewReader(encodedPayload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build thrift request: %w", err)
	}

	for _, header := range headers {
		request.Header.Add(header.Key, header.Value)
	}

	responseBody, err := f.executeRequest(request)
	if err != nil {
		return "", err
	}

	thriftResponse, err := thrift.ResponseBytesToMap(
		function.Spec(),
		responseBody,
		thrift.Options{UseEnvelopes: true},
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse thrift response: %w", err)
	}

	jsonResponse, err := json.Marshal(thriftResponse)
	if err != nil {
		return "", fmt.Errorf("failed to marshal a response: %w", err)
	}

	return string(jsonResponse), nil
}

func (f *Form) StopCurrentRequest() {
	if f.requestCancelFunc == nil {
		return
	}

	f.requestCancelFunc()
	f.requestCancelFunc = nil
}

func (f *Form) Close() error {
	f.client.CloseIdleConnections()

	return nil
}

func (f *Form) executeRequest(request *http.Request) (_ []byte, rerr error) {
	response, err := f.client.Do(request)
	defer func() {
		if response == nil {
			return
		}

		err := response.Body.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a response body: %w", err))
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return responseBody, nil
}
