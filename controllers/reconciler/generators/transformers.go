package generators

import (
	"reflect"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/imdario/mergo"
)

type transformResult struct {
	Params   []map[string]string
	Template sourcev1.KustomizationSetTemplate
}

// Transform a spec generator to list of params and a template
func Transform(generator sourcev1.KustomizationSetGenerator, allGenerators map[string]Generator, baseTemplate sourcev1.KustomizationSetTemplate, kustomizeSet *sourcev1.KustomizationSet) ([]transformResult, error) {
	res := []transformResult{}
	var firstError error

	generators := findRelevantGenerators(&generator, allGenerators)
	for _, g := range generators {
		mergedTemplate, err := mergeGeneratorTemplate(g, &generator, baseTemplate)
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			continue
		}

		params, err := g.GenerateParams(&generator, kustomizeSet)
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			continue
		}

		res = append(res, transformResult{
			Params:   params,
			Template: mergedTemplate,
		})
	}
	return res, firstError
}

func findRelevantGenerators(requestedGenerator *sourcev1.KustomizationSetGenerator, generators map[string]Generator) []Generator {
	var res []Generator
	v := reflect.Indirect(reflect.ValueOf(requestedGenerator))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}

		if !reflect.ValueOf(field.Interface()).IsNil() {
			res = append(res, generators[v.Type().Field(i).Name])
		}
	}
	return res
}

func mergeGeneratorTemplate(g Generator, requestedGenerator *sourcev1.KustomizationSetGenerator, kustomizationSetTemplate sourcev1.KustomizationSetTemplate) (sourcev1.KustomizationSetTemplate, error) {
	// Make a copy of the value from `GetTemplate()` before merge, rather than copying directly into
	// the provided parameter (which will touch the original resource object returned by client-go)
	dest := g.GetTemplate(requestedGenerator).DeepCopy()
	if dest == nil {
		return kustomizationSetTemplate, nil
	}
	err := mergo.Merge(dest, kustomizationSetTemplate)
	return *dest, err
}
