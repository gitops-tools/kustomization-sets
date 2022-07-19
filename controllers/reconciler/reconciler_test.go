package reconciler

import (
	"context"
	"testing"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/controllers/reconciler/generators"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testKustomizationSetName = "test-kustomizations"

func TestGenerateKustomizations(t *testing.T) {
	testGenerators := map[string]generators.Generator{
		"List": generators.NewListGenerator(),
	}
	listGeneratorTests := []struct {
		name     string
		elements []apiextensionsv1.JSON
		tp       *sourcev1.KustomizationSetTemplate
		want     []kustomizev1.Kustomization
	}{
		{
			name: "multiple elements",
			elements: []apiextensionsv1.JSON{
				{Raw: []byte(`{"cluster": "engineering-dev"}`)},
				{Raw: []byte(`{"cluster": "engineering-prod"}`)},
				{Raw: []byte(`{"cluster": "engineering-preprod"}`)},
			},
			want: []kustomizev1.Kustomization{
				makeTestKustomization("engineering-dev"),
				makeTestKustomization("engineering-prod"),
				makeTestKustomization("engineering-preprod"),
			},
		},
		{
			name: "list generator with template",
			elements: []apiextensionsv1.JSON{
				{Raw: []byte(`{"cluster": "engineering-dev"}`)},
			},
			tp: &sourcev1.KustomizationSetTemplate{
				KustomizationSetTemplateMeta: sourcev1.KustomizationSetTemplateMeta{
					Labels: map[string]string{
						"cluster.app/name": "{{ cluster }}",
					},
				},
			},
			want: []kustomizev1.Kustomization{
				makeTestKustomization("engineering-dev", withLabels(map[string]string{"cluster.app/name": "engineering-dev"})),
			},
		},
	}

	for _, tt := range listGeneratorTests {
		t.Run(tt.name, func(t *testing.T) {
			kset := makeTestKustomizationSet(withListElements(tt.elements, tt.tp))
			kusts, err := GenerateKustomizations(context.TODO(), kset, testGenerators)
			test.AssertNoError(t, err)

			if diff := cmp.Diff(tt.want, kusts); diff != "" {
				t.Fatalf("failed to generate kustomizations:\n%s", diff)
			}
		})
	}
}

func withListElements(el []apiextensionsv1.JSON, tp *sourcev1.KustomizationSetTemplate) func(*sourcev1.KustomizationSet) {
	return func(ks *sourcev1.KustomizationSet) {
		if ks.Spec.Generators == nil {
			ks.Spec.Generators = []sourcev1.KustomizationSetGenerator{}
		}
		ks.Spec.Generators = append(ks.Spec.Generators,
			sourcev1.KustomizationSetGenerator{

				List: &sourcev1.ListGenerator{
					Template: tp,
					Elements: el,
				},
			})
	}
}

func makeTestKustomizationSet(opts ...func(*sourcev1.KustomizationSet)) *sourcev1.KustomizationSet {
	ks := &sourcev1.KustomizationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: testKustomizationSetName,
		},
		Spec: sourcev1.KustomizationSetSpec{
			Template: sourcev1.KustomizationSetTemplate{
				KustomizationSetTemplateMeta: sourcev1.KustomizationSetTemplateMeta{
					Name: `{{cluster}}-demo`,
				},
				Spec: kustomizev1.KustomizationSpec{
					Interval: metav1.Duration{Duration: 5 * time.Minute},
					Path:     "./clusters/{{cluster}}/",
					Prune:    true,
					SourceRef: kustomizev1.CrossNamespaceSourceReference{
						Kind: "GitRepository",
						Name: "demo-repo",
					},
					KubeConfig: &kustomizev1.KubeConfig{
						SecretRef: meta.SecretKeyReference{
							Name: "{{cluster}}",
						},
					},
				},
			},
		},
	}
	for _, o := range opts {
		o(ks)
	}
	return ks
}

func withLabels(l map[string]string) func(*kustomizev1.Kustomization) {
	return func(k *kustomizev1.Kustomization) {
		k.ObjectMeta.Labels = l
	}
}

func makeTestKustomization(name string, opts ...func(*kustomizev1.Kustomization)) kustomizev1.Kustomization {
	k := kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-demo",
		},
		Spec: kustomizev1.KustomizationSpec{
			Path:     "./clusters/" + name + "/",
			Interval: metav1.Duration{Duration: 5 * time.Minute},
			Prune:    true,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: "demo-repo",
			},
			KubeConfig: &kustomizev1.KubeConfig{
				SecretRef: meta.SecretKeyReference{
					Name: name,
				},
			},
		},
	}
	for _, o := range opts {
		o(&k)
	}
	return k
}
