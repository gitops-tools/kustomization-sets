package generators

import (
	"testing"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
)

func TestTransform(t *testing.T) {
	transformTests := []struct {
		name     string
		gen      sourcev1.KustomizationSetGenerator
		gens     map[string]Generator
		template sourcev1.KustomizationSetTemplate
		set      *sourcev1.KustomizationSet
	}{{}}
}

func TestTransform_errors(t *testing.T) {
	t.Skip()
}
