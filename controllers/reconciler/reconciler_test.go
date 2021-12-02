package reconciler

import (
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

const testKustomizationSetName = "test-kustomizations"

func TestGenerateKustomizations(t *testing.T) {
	kset := makeTestKustomizationSet()

	kusts, err := GenerateKustomizations(kset)
	assert.NoError(t, err)

	want := []kustomizev1.Kustomization{
		makeTestKustomization("engineering-dev"),
		makeTestKustomization("engineering-prod"),
		makeTestKustomization("engineering-preprod"),
	}
	if diff := cmp.Diff(want, kusts); diff != "" {
		t.Fatalf("failed to generate kustomizations:\n%s", diff)
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
				KustomizationSetTemplateMeta: sourcev1.KustomizationSetTemplateMeta{
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

func makeTestKustomization(name string) kustomizev1.Kustomization {
	return kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-demo",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path:     "./clusters/" + name + "/",
			Interval: metav1.Duration{5 * time.Minute},
			Prune:    true,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: "demo-repo",
			},
			KubeConfig: &kustomizev1.KubeConfig{
				SecretRef: meta.LocalObjectReference{
					Name: name,
				},
			},
		},
	}
}
