package grpc

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/fullstorydev/grpcurl"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/samber/lo"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

type Project struct {
	ID             string           `json:"id"`
	SplitterWidth  float64          `json:"splitterWidth"`
	Forms          map[string]*Form `json:"forms"`
	FormIDs        []string         `json:"formIDs"`
	CurrentFormID  string           `json:"currentFormID"`
	IsReflected    bool             `json:"isReflected"`
	ImportPathList []string         `json:"importPathList"`
	ProtoFileList  []string         `json:"protoFileList"`
	Nodes          []*ProtoTreeNode `json:"nodes"`

	stateMutex            sync.RWMutex
	stateStorage          *state.Storage
	protoTree             *ProtoTree
	protoDescriptorSource grpcurl.DescriptorSource
}

func NewProject(projectID string, stateStorage *state.Storage) (*Project, error) {
	formID := uuid.Must(uuid.NewV4()).String()
	address := "0.0.0.0:50051"

	project := &Project{
		ID:            projectID,
		SplitterWidth: defaultProjectSplitterWidth,
		Forms: map[string]*Form{
			formID: {
				ID:       formID,
				Address:  address,
				Request:  "{}",
				Response: "{}",
			},
		},
		CurrentFormID: formID,
		stateStorage:  stateStorage,
	}
	project.FormIDs = append(project.FormIDs, formID)

	if err := project.saveState(); err != nil {
		return nil, err
	}

	return project, nil
}

func (p *Project) IsProtoDescriptorSourceInitialized() bool {
	return p.protoDescriptorSource != nil
}

func (p *Project) SendRequest(
	formID,
	methodID,
	address,
	payload string,
	headers []*Header,
) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Address = address
	p.Forms[formID].Request = payload

	if p.IsReflected && !p.IsProtoDescriptorSourceInitialized() {
		_, err := p.reflectProto(formID, address)
		if err != nil {
			return err
		}
	}

	form := p.Forms[formID]

	response, err := form.SendRequest(methodID, address, payload, p.protoDescriptorSource, headers)
	if err != nil {
		p.Forms[formID].Response = "{}"

		return err
	}

	p.Forms[formID].Response = response

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) StopRequest(id string) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	form := p.Forms[id]

	form.StopCurrentRequest()
}

func (p *Project) ReflectProto(formID, address string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	nodes, err := p.reflectProto(formID, address)
	if err != nil {
		return err
	}

	form := p.Forms[p.CurrentFormID]
	form.SelectedMethodID = ""
	form.Request = "{}"
	form.Response = "{}"

	p.IsReflected = true
	p.Nodes = nodes
	p.ImportPathList = nil
	p.ProtoFileList = nil

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) RemoveImportPath(importPath string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.ImportPathList = lo.Reject(
		p.ImportPathList,
		func(ip string, _ int) bool {
			return ip == importPath
		},
	)

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) OpenProtoFile(protoFilePath string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if lo.Contains(p.ProtoFileList, protoFilePath) {
		return nil
	}

	var importPathList []string
	if len(p.ImportPathList) > 0 {
		importPathList = p.ImportPathList
	} else {
		currentDir := path.Dir(protoFilePath)
		importPathList = []string{currentDir}
	}

	protoFileList := append([]string{protoFilePath}, p.ProtoFileList...)

	nodes, err := p.RefreshProtoDescriptors(importPathList, protoFileList)
	if err != nil {
		return err
	}

	p.IsReflected = false
	p.Nodes = nodes
	p.ImportPathList = importPathList
	p.ProtoFileList = protoFileList

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) DeleteAllProtoFiles() error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.IsReflected = false
	p.ProtoFileList = nil

	nodes, err := p.RefreshProtoDescriptors(
		p.ImportPathList,
		p.ProtoFileList,
	)
	if err != nil {
		return err
	}

	p.Nodes = nodes

	for _, form := range p.Forms {
		if form.ID == p.CurrentFormID {
			continue
		}

		err := form.Close()
		if err != nil {
			return err
		}
	}

	form := p.Forms[p.CurrentFormID]
	form.SelectedMethodID = ""
	form.Request = "{}"
	form.Response = "{}"

	p.Forms = map[string]*Form{form.ID: form}

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) OpenImportPath(importPath string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if lo.Contains(p.ImportPathList, importPath) {
		return nil
	}

	p.ImportPathList = append(p.ImportPathList, importPath)

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) SelectMethod(methodID, formID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	method := p.protoTree.Method(methodID)
	methodMessage := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(method.Descriptor().GetInputType())

	methodPayloadJSON, err := methodMessage.MarshalJSONPB(&jsonpb.Marshaler{EmitDefaults: true, OrigName: true})
	if err != nil {
		return fmt.Errorf("failed to prepare grpc request: %w", err)
	}

	payloadJSONStr := string(methodPayloadJSON)

	formattedJSON, err := jsbeautifier.Beautify(&payloadJSONStr, jsbeautifier.DefaultOptions())
	if err != nil {
		return fmt.Errorf("failed to format a method payload: %w", err)
	}

	p.Forms[formID].Request = formattedJSON
	p.Forms[formID].Response = "{}"
	p.Forms[formID].SelectedMethodID = methodID

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveCurrentFormID(currentFormID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.CurrentFormID = currentFormID

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveAddress(formID, address string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Address = address

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) AddHeader(formID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Headers = append(
		p.Forms[formID].Headers,
		&Header{
			ID:    uuid.Must(uuid.NewV4()).String(),
			Key:   "",
			Value: "",
		},
	)

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveHeaders(formID string, headers []*Header) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Headers = headers

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) DeleteHeader(formID, headerID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Headers = lo.Reject(
		p.Forms[formID].Headers,
		func(header *Header, _ int) bool {
			return header.ID == headerID
		},
	)

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveSplitterWidth(splitterWidth float64) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.SplitterWidth = splitterWidth

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveRequestPayload(formID, requestPayload string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Request = requestPayload

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) CreateNewForm() error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	var headers []*Header

	address := "0.0.0.0:50051"
	if p.CurrentFormID != "" {
		address = p.Forms[p.CurrentFormID].Address
		headers = p.Forms[p.CurrentFormID].Headers
	}

	p.Forms[formID] = &Form{
		ID:       formID,
		Address:  address,
		Request:  "{}",
		Response: "{}",
		Headers:  headers,
	}
	p.FormIDs = append(p.FormIDs, formID)
	p.CurrentFormID = formID

	if err := p.saveState(); err != nil {
		return err
	}

	return nil
}

func (p *Project) RemoveForm(formID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if len(p.Forms) <= 1 {
		return nil
	}

	delete(p.Forms, formID)
	p.FormIDs = lo.Reject(
		p.FormIDs,
		func(fID string, _ int) bool {
			return formID == fID
		},
	)

	if p.CurrentFormID == formID {
		p.CurrentFormID = lo.Keys(p.Forms)[0]
	}

	if err := p.Forms[formID].Close(); err != nil {
		return err
	}

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

	return p.refreshProtoNodes(protoDescriptorSource)
}

func (p *Project) Close() error {
	for _, form := range p.Forms {
		err := form.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) refreshProtoNodes(protoDescriptorSource grpcurl.DescriptorSource) ([]*ProtoTreeNode, error) {
	protoTree, err := NewProtoTree(protoDescriptorSource)
	if err != nil {
		return nil, err
	}

	p.protoDescriptorSource = protoDescriptorSource
	p.protoTree = protoTree

	return protoTree.Nodes(), nil
}

func (p *Project) reflectProto(formID, address string) ([]*ProtoTreeNode, error) {
	form := p.Forms[formID]

	protoDescriptorSource, err := form.ReflectProto(context.Background(), address)
	if err != nil {
		return nil, err
	}

	nodes, err := p.refreshProtoNodes(protoDescriptorSource)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (p *Project) saveState() error {
	err := p.stateStorage.Save(p.ID, p)
	if err != nil {
		return fmt.Errorf("failed to store a grpc project: %w", err)
	}

	return nil
}
