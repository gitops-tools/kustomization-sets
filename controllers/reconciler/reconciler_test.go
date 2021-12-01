package reconciler

import (
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateKustomizations parses the KustomizationSet and creates a
// Kustomization using the configured generators and templates.
func GenerateKustomizations(r *sourcev1.KustomizationSet) ([]*kustomizev1.Kustomization, error) {
	return nil, nil
}

const testKustomizationSetName = "test-kustomizations"

func TestGenerateKustomizations(t *testing.T) {
	kset := makeTestKustomizationSet()

	kusts, err := GenerateKustomizations(kset)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(kusts); l != 3 {
		t.Fatalf("got %d, want 3", l)
	}
}

func makeTestKustomizationSet() *sourcev1.KustomizationSet {
	return &sourcev1.KustomizationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: testKustomizationSetName,
		},
		Spec: sourcev1.KustomizationSetSpec{
			Generators: []sourcev1.KustomizationSetGenerator{
				{
					List: &sourcev1.ListGenerator{
						Elements: []apiextensionsv1.JSON{
							{Raw: []byte(`{"cluster": "engineering-dev"}`)},
							{Raw: []byte(`{"cluster": "engineering-prod"}`)},
							{Raw: []byte(`{"cluster": "engineering-preprod"}`)},
						},
					},
				},
			},
			Template: sourcev1.KustomizationSetTemplate{
				Metadata: sourcev1.KustomizationSetTemplateMeta{
					Name: `{{cluster}}-demo`,
				},
				Spec: kustomizev1.KustomizationSpec{
					Interval: metav1.Duration{5 * time.Minute},
					Path:     "./clusters/{{cluster}}/",
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind: "GitRepository",
						Name: "demo-repo",
					},
					KubeConfig: &kustomizev1.KubeConfig{
						SecretRef: meta.LocalObjectReference{
							Name: "{{cluster}}",
						},
					},
				},
			},
		},
	}
}
