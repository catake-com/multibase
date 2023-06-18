package thrift

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"github.com/gofrs/uuid/v5"
	"github.com/samber/lo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.uber.org/thriftrw/compile"

	"github.com/catake-com/multibase/backend/pkg/state"
)

var errThriftUnknownType = errors.New("unknown type during thrift parsing")

type Project struct {
	ID            string             `json:"id"`
	SplitterWidth float64            `json:"splitterWidth"`
	Forms         map[string]*Form   `json:"forms"`
	FormIDs       []string           `json:"formIDs"`
	CurrentFormID string             `json:"currentFormID"`
	FilePath      string             `json:"filePath"`
	Nodes         []*ServiceTreeNode `json:"nodes"`

	stateMutex   sync.RWMutex
	stateStorage *state.Storage
	serviceTree  *ServiceTree
}

func NewProject(projectID string, stateStorage *state.Storage) (*Project, error) {
	formID := uuid.Must(uuid.NewV4()).String()
	address := "0.0.0.0:9090"

	project := &Project{
		ID:            projectID,
		SplitterWidth: defaultProjectSplitterWidth,
		Forms: map[string]*Form{
			formID: {
				ID:            formID,
				Address:       address,
				IsMultiplexed: true,
				Request:       "{}",
				Response:      "{}",
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

func (p *Project) CreateNewForm() error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	formID := uuid.Must(uuid.NewV4()).String()

	var headers []*Header

	address := "0.0.0.0:9090"
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

	return p.saveState()
}

func (p *Project) RemoveForm(formID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if len(p.Forms) <= 1 {
		return nil
	}

	form := p.Forms[formID]

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

	if err := form.Close(); err != nil {
		return err
	}

	return p.saveState()
}

func (p *Project) SendRequest(
	formID,
	address,
	payload string,
) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Address = address
	p.Forms[formID].Request = payload

	response, err := p.Forms[formID].SendRequest(
		p.Forms[formID].SelectedFunctionID,
		address,
		payload,
		p.Forms[formID].IsMultiplexed,
		p.Forms[formID].Headers,
	)
	if err != nil {
		p.Forms[formID].Response = "{}"

		return err
	}

	p.Forms[formID].Response = response

	return p.saveState()
}

func (p *Project) StopRequest(formID string) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	form := p.Forms[formID]

	form.StopCurrentRequest()
}

func (p *Project) OpenFilePath(filePath string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	nodes, err := p.generateNodes(filePath)
	if err != nil {
		return err
	}

	if p.FilePath != "" {
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
		form.SelectedFunctionID = ""
		form.Request = "{}"
		form.Response = "{}"

		p.Forms = map[string]*Form{form.ID: form}
	}

	p.Nodes = nodes
	p.FilePath = filePath

	return p.saveState()
}

func (p *Project) GenerateServiceTreeNodes(filePath string) ([]*ServiceTreeNode, error) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	serviceTreeNodes, err := p.generateNodes(filePath)
	if err != nil {
		return nil, err
	}

	return serviceTreeNodes, nil
}

func (p *Project) SelectFunction(formID, functionID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	function := p.serviceTree.Function(functionID)
	payload := orderedmap.New[string, interface{}]()

	for _, argsSpec := range function.Spec().ArgsSpec {
		v, err := parseThriftType(argsSpec.Type)
		if err != nil {
			return err
		}

		payload.Set(argsSpec.Name, v)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal a function payload: %w", err)
	}

	payloadJSONStr := string(payloadJSON)

	formattedJSON, err := jsbeautifier.Beautify(&payloadJSONStr, jsbeautifier.DefaultOptions())
	if err != nil {
		return fmt.Errorf("failed to format a function payload: %w", err)
	}

	p.Forms[formID].Request = formattedJSON
	p.Forms[formID].Response = "{}"
	p.Forms[formID].SelectedFunctionID = functionID

	return p.saveState()
}

func (p *Project) SaveCurrentFormID(currentFormID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.CurrentFormID = currentFormID

	return p.saveState()
}

func (p *Project) SaveAddress(formID, address string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Address = address

	return p.saveState()
}

func (p *Project) SaveIsMultiplexed(formID string, isMultiplexed bool) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].IsMultiplexed = isMultiplexed

	return p.saveState()
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

	return p.saveState()
}

func (p *Project) SaveHeaders(formID string, headers []*Header) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Headers = headers

	return p.saveState()
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

	return p.saveState()
}

func (p *Project) SaveSplitterWidth(splitterWidth float64) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.SplitterWidth = splitterWidth

	return p.saveState()
}

func (p *Project) SaveRequestPayload(formID, requestPayload string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	p.Forms[formID].Request = requestPayload

	return p.saveState()
}

func (p *Project) BeautifyRequest(formID string) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	formattedJSON, err := jsbeautifier.Beautify(&p.Forms[formID].Request, jsbeautifier.DefaultOptions())
	if err != nil {
		return nil // nolint: nilerr
	}

	p.Forms[formID].Request = formattedJSON

	return p.saveState()
}

func (p *Project) Close() error {
	for _, client := range p.Forms {
		err := client.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) saveState() error {
	err := p.stateStorage.Save(p.ID, p)
	if err != nil {
		return fmt.Errorf("failed to store a thrift project: %w", err)
	}

	return nil
}

func (p *Project) generateNodes(filePath string) ([]*ServiceTreeNode, error) {
	module, err := compile.Compile(filePath, compile.NonStrict())
	if err != nil {
		return nil, fmt.Errorf("failed compile thrift: %w", err)
	}

	serviceTree, err := NewServiceTree(module)
	if err != nil {
		return nil, err
	}

	p.serviceTree = serviceTree

	for _, form := range p.Forms {
		form.serviceTree = serviceTree
	}

	return serviceTree.Nodes(), nil
}

// nolint: funlen, cyclop
func parseThriftType(typ compile.TypeSpec) (interface{}, error) {
	switch spec := typ.(type) {
	case *compile.StructSpec:
		str := orderedmap.New[string, interface{}]()

		for _, field := range spec.Fields {
			v, err := parseThriftType(field.Type)
			if err != nil {
				return nil, err
			}

			str.Set(field.Name, v)
		}

		return str, nil
	case *compile.TypedefSpec:
		return parseThriftType(spec.Target)
	case *compile.StringSpec:
		return "", nil
	case *compile.BoolSpec:
		return false, nil
	case *compile.I8Spec, *compile.I16Spec, *compile.I32Spec, *compile.I64Spec, *compile.EnumSpec:
		return 0, nil
	case *compile.BinarySpec:
		return []byte{}, nil
	case *compile.DoubleSpec:
		return 0.0, nil
	case *compile.ListSpec:
		v, err := parseThriftType(spec.ValueSpec)
		if err != nil {
			return nil, err
		}

		return []interface{}{v}, nil
	case *compile.SetSpec:
		v, err := parseThriftType(spec.ValueSpec)
		if err != nil {
			return nil, err
		}

		return []interface{}{v}, nil
	case *compile.MapSpec:
		key, err := parseThriftType(spec.KeySpec)
		if err != nil {
			return nil, err
		}

		value, err := parseThriftType(spec.ValueSpec)
		if err != nil {
			return nil, err
		}

		return map[interface{}]interface{}{key: value}, nil
	default:
		return nil, fmt.Errorf("failed to parse %v: %w", typ, errThriftUnknownType)
	}
}
