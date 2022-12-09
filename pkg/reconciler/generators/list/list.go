package list

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sourcev1 "github.com/gitops-tools/kustomization-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomization-set-controller/pkg/reconciler/generators"
)

// ListGenerator is a generic JSON object list.
type ListGenerator struct {
}

// NewGenerator creates and returns a new list generator.
func NewGenerator() *ListGenerator {
	return &ListGenerator{}
}

func (g *ListGenerator) Generate(_ context.Context, sg *sourcev1.KustomizationSetGenerator, _ *sourcev1.KustomizationSet) ([]map[string]any, error) {
	if sg == nil {
		return nil, generators.EmptyKustomizationSetGeneratorError
	}

	if sg.List == nil {
		return nil, nil
	}

	res := make([]map[string]any, len(sg.List.Elements))
	for i, el := range sg.List.Elements {
		element := map[string]any{}
		if err := json.Unmarshal(el.Raw, &element); err != nil {
			return nil, fmt.Errorf("error unmarshaling list element: %w", err)
		}
		res[i] = element
	}

	return res, nil
}

// Interval is an implementation of the Generator interface.
func (g *ListGenerator) Interval(sg *sourcev1.KustomizationSetGenerator) time.Duration {
	return generators.NoRequeueInterval
}

// Template is an implementation of the Generator interface.
func (g *ListGenerator) Template(sg *sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate {
	return sg.List.Template
}
