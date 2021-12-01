package reconciler

import (
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/controllers/reconciler/generators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultGenerators = map[string]generators.Generator{
	"List": generators.NewListGenerator(),
}

// GenerateKustomizations parses the KustomizationSet and creates a
// Kustomization using the configured generators and templates.
func GenerateKustomizations(r *sourcev1.KustomizationSet) ([]kustomizev1.Kustomization, error) {
	var res []kustomizev1.Kustomization
	var firstError error

	for _, g := range r.Spec.Generators {
		t, err := generators.Transform(g, defaultGenerators, r.Spec.Template, r)
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			continue
		}

		for _, a := range t {
			tmplApplication := makeKustomization(a.Template)

			for _, p := range a.Params {
				app, err := defaultRenderer.RenderTemplateParams(tmplApplication, p)
				if err != nil {
					if firstError == nil {
						firstError = err
					}
					continue
				}
				res = append(res, *app)
			}
		}
	}
	return res, firstError
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
