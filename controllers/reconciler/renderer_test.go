package reconciler

import (
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRenderTemplate(t *testing.T) {
	newKustomization := func(opts ...func(*kustomizev1.Kustomization)) *kustomizev1.Kustomization {
		k := &kustomizev1.Kustomization{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testing",
			},
			Spec: kustomizev1.KustomizationSpec{
				Path:     "testing",
				Interval: metav1.Duration{Duration: 5 * time.Minute},
				Prune:    true,
				SourceRef: kustomizev1.CrossNamespaceSourceReference{
					Kind: "GitRepository",
					Name: "testing",
				},
			},
		}
		for _, o := range opts {
			o(k)
		}
		return k
	}

	templatePath := func(s string) func(*kustomizev1.Kustomization) {
		return func(k *kustomizev1.Kustomization) {
			k.Spec.Path = s
		}
	}

	templateTests := []struct {
		name   string
		tmpl   *kustomizev1.Kustomization
		params map[string]string
		want   *kustomizev1.Kustomization
	}{
		{name: "no params", tmpl: newKustomization(), want: newKustomization()},
		{
			name:   "simple params",
			tmpl:   newKustomization(templatePath("{{.replaced}}")),
			params: map[string]string{"replaced": "new string"},
			want:   newKustomization(templatePath("new string")),
		},
		{
			name:   "sanitize",
			tmpl:   newKustomization(templatePath("{{ sanitize .replaced }}")),
			params: map[string]string{"replaced": "new string"},
			want:   newKustomization(templatePath("newstring")),
		},
	}

	for _, tt := range templateTests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := renderTemplateParams(tt.tmpl, tt.params)
			test.AssertNoError(t, err)

			if diff := cmp.Diff(tt.want, rendered); diff != "" {
				t.Fatalf("rendering failed:\n%s", diff)
			}
		})
	}

}

func TestRenderTemplate_errors(t *testing.T) {
	templateTests := []struct {
		name    string
		tmpl    *kustomizev1.Kustomization
		params  map[string]string
		wantErr string
	}{
		{name: "no template", tmpl: nil, params: nil, wantErr: "template is empty"},
	}

	for _, tt := range templateTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := renderTemplateParams(tt.tmpl, tt.params)

			test.AssertErrorMatch(t, tt.wantErr, err)
		})
	}
}
