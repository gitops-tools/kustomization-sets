package generators

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
)

// ListGenerator is a generic JSON object list.
type ListGenerator struct {
}

// NewListGenerator creates and returns a new list generator.
func NewListGenerator() Generator {
	return &ListGenerator{}
}

func (g *ListGenerator) GenerateParams(_ context.Context, sg *sourcev1.KustomizationSetGenerator, _ *sourcev1.KustomizationSet) ([]map[string]string, error) {
	if sg == nil {
		return nil, EmptyKustomizationSetGeneratorError
	}

	if sg.List == nil {
		return nil, nil
	}

	res := make([]map[string]string, len(sg.List.Elements))
	for i, el := range sg.List.Elements {
		params := map[string]string{}
		var element map[string]interface{}
		err := json.Unmarshal(el.Raw, &element)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling list element: %w", err)
		}

		for key, value := range element {
			if key == "values" {
				values, ok := (value).(map[string]interface{})
				if !ok {
					return nil, errors.New("error parsing values map")
				}
				for k, v := range values {
					value, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("error parsing value as string %w", err)
					}
					params[fmt.Sprintf("values.%s", k)] = value
				}
			} else {
				v, ok := value.(string)
				if !ok {
					return nil, fmt.Errorf("error parsing value as string %w", err)
				}
				params[key] = v
			}
		}
		res[i] = params
	}
	return res, nil
}

// GetInterval is an implementation of the Generator interface.
func (g *ListGenerator) GetInterval(sg *sourcev1.KustomizationSetGenerator) time.Duration {
	return NoRequeueInterval
}

// GetTemplate is an implementation of the Generator interface.
func (g *ListGenerator) GetTemplate(sg *sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate {
	return sg.List.Template
}
