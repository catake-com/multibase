package thrift

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
	"go.uber.org/thriftrw/compile"
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

	serviceTree *ServiceTree
}

func (p *Project) SendRequest(
	formID,
	address,
	functionID,
	payload string,
	isMultiplexed bool,
	headers []*Header,
) (string, error) {
	form := p.Forms[formID]

	return form.SendRequest(functionID, address, payload, isMultiplexed, headers)
}

func (p *Project) StopRequest(formID string) {
	form := p.Forms[formID]

	form.StopCurrentRequest()
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

	for _, form := range p.Forms {
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

	payloadJSONStr := string(payloadJSON)

	formattedJSON, err := jsbeautifier.Beautify(&payloadJSONStr, jsbeautifier.DefaultOptions())
	if err != nil {
		return "", fmt.Errorf("failed to format a function payload: %w", err)
	}

	return formattedJSON, nil
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
