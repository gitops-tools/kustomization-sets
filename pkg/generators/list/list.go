package list

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sourcev1 "github.com/gitops-tools/kustomization-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomization-set-controller/pkg/generators"
	"github.com/go-logr/logr"
)

// ListGenerator is a generic JSON object list.
type ListGenerator struct {
	logger logr.Logger
}

// GeneratorFactory is a function for creating per-reconciliation generators the
// ListGenerator.
func GeneratorFactory() generators.GeneratorFactory {
	return func(l logr.Logger) generators.Generator {
		return NewGenerator(l)
	}
}

// NewGenerator creates and returns a new list generator.
func NewGenerator(l logr.Logger) *ListGenerator {
	return &ListGenerator{logger: l}
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
