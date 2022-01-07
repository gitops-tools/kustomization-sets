package reconciler

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/controllers/reconciler/generators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var knownGenerators = map[string]generators.Generator{
	"List": generators.NewListGenerator(),
}

// GenerateKustomizations parses the KustomizationSet and creates a
// Kustomization using the configured generators and templates.
func GenerateKustomizations(ctx context.Context, r *sourcev1.KustomizationSet) ([]kustomizev1.Kustomization, error) {
	var res []kustomizev1.Kustomization
	for _, g := range r.Spec.Generators {
		t, err := transform(ctx, g, knownGenerators, r.Spec.Template, r)
		if err != nil {
			return nil, fmt.Errorf("failed to transform template for set %s: %w", r.GetName(), err)
		}
		for _, a := range t {
			tmplKustomization := makeKustomization(a.Template)
			for _, p := range a.Params {
				app, err := renderTemplateParams(tmplKustomization, p)
				if err != nil {
					return nil, fmt.Errorf("failed to render template params for set %s: %w", r.GetName(), err)
				}
				res = append(res, *app)
			}
		}
	}
	return res, nil
}

func makeKustomization(template sourcev1.KustomizationSetTemplate) *kustomizev1.Kustomization {
	return &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: template.Annotations,
			Labels:      template.Labels,
			Namespace:   template.Namespace,
			Name:        template.Name,
			Finalizers:  template.Finalizers,
		},
		Spec: template.Spec,
	}
}
