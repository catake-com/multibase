package thrift

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/thriftrw/compile"
)

var (
	errThriftUnknownType = errors.New("unknown type during thrift parsing")
)

type Project struct {
	id          string
	forms       map[string]*Form
	serviceTree *ServiceTree
}

func NewProject(id string) *Project {
	return &Project{
		id:    id,
		forms: make(map[string]*Form),
	}
}

func (p *Project) SendRequest(formID, address, functionID, payload string) (string, error) {
	form := p.forms[formID]

	return form.SendRequest(functionID, address, payload)
}

func (p *Project) StopRequest(formID string) {
	form := p.forms[formID]

	form.StopCurrentRequest()
}

func (p *Project) InitializeForm(formID string) error {
	form, err := NewForm(formID, p.serviceTree)
	if err != nil {
		return err
	}

	p.forms[formID] = form

	return nil
}

func (p *Project) GenerateServiceTreeNodes(filePath string) ([]*ServiceTreeNode, error) {
	module, err := compile.Compile(filePath, compile.NonStrict())
	if err != nil {
		return nil, fmt.Errorf("failed compile thrift: %w", err)
	}

	serviceTree, err := NewServiceTree(module)
	if err != nil {
		return nil, err
	}

	p.serviceTree = serviceTree

	for _, form := range p.forms {
		form.serviceTree = serviceTree
	}

	return serviceTree.Nodes(), nil
}

func (p *Project) SelectFunction(functionID string) (string, error) {
	function := p.serviceTree.Function(functionID)
	payload := make(map[string]interface{})

	for _, argsSpec := range function.Spec().ArgsSpec {
		v, err := parseThriftType(argsSpec.Type)
		if err != nil {
			return "", err
		}

		payload[argsSpec.Name] = v
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal a function payload: %w", err)
	}

	return string(payloadJSON), nil
}

func (p *Project) Close() error {
	for _, client := range p.forms {
		err := client.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// nolint: funlen, cyclop
func parseThriftType(typ compile.TypeSpec) (interface{}, error) {
	switch spec := typ.(type) {
	case *compile.StructSpec:
		str := map[string]interface{}{}

		for _, field := range spec.Fields {
			v, err := parseThriftType(field.Type)
			if err != nil {
				return nil, err
			}

			str[field.Name] = v
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
		k, err := parseThriftType(spec.KeySpec)
		if err != nil {
			return nil, err
		}

		v, err := parseThriftType(spec.ValueSpec)
		if err != nil {
			return nil, err
		}

		return map[interface{}]interface{}{k: v}, nil
	default:
		return nil, fmt.Errorf("failed to parse %v: %w", typ, errThriftUnknownType)
	}
}
