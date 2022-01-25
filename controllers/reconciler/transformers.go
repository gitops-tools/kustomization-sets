package reconciler

import (
	"context"
	"reflect"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/controllers/reconciler/generators"
	"github.com/imdario/mergo"
)

type transformResult struct {
	Params   []map[string]string
	Template sourcev1.KustomizationSetTemplate
}

func transform(ctx context.Context, generator sourcev1.KustomizationSetGenerator, allGenerators map[string]generators.Generator, baseTemplate sourcev1.KustomizationSetTemplate, kustomizeSet *sourcev1.KustomizationSet) ([]transformResult, error) {
	res := []transformResult{}
	generators := findRelevantGenerators(&generator, allGenerators)
	for _, g := range generators {
		mergedTemplate, err := mergeGeneratorTemplate(g, &generator, baseTemplate)
		if err != nil {
			return nil, err
		}

		params, err := g.Generate(ctx, &generator, kustomizeSet)
		if err != nil {
			return nil, err
		}

		res = append(res, transformResult{
			Params:   params,
			Template: mergedTemplate,
		})
	}
	return res, nil
}

func findRelevantGenerators(setGenerator *sourcev1.KustomizationSetGenerator, allGenerators map[string]generators.Generator) []generators.Generator {
	var res []generators.Generator
	v := reflect.Indirect(reflect.ValueOf(setGenerator))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}

		if !reflect.ValueOf(field.Interface()).IsNil() {
			res = append(res, allGenerators[v.Type().Field(i).Name])
		}
	}
	return res
}

func mergeGeneratorTemplate(g generators.Generator, setGenerator *sourcev1.KustomizationSetGenerator, kustomizationSetTemplate sourcev1.KustomizationSetTemplate) (sourcev1.KustomizationSetTemplate, error) {
	// Make a copy of the value from `Template()` before merge, rather than copying directly into
	// the provided parameter (which will touch the original resource object returned by client-go)
	dest := g.Template(setGenerator).DeepCopy()
	if dest == nil {
		return kustomizationSetTemplate, nil
	}
	err := mergo.Merge(dest, kustomizationSetTemplate)
	return *dest, err
}
