package generators

import (
	"context"
	"testing"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ Generator = (*ListGenerator)(nil)

func TestGenerateListParams(t *testing.T) {
	testCases := []struct {
		elements []apiextensionsv1.JSON
		want     []map[string]string
	}{
		{
			elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url"}`)}},
			want:     []map[string]string{{"cluster": "cluster", "url": "url"}},
		}, {
			elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url","values":{"foo":"bar"}}`)}},
			want:     []map[string]string{{"cluster": "cluster", "url": "url", "values.foo": "bar"}},
		},
	}

	for _, tt := range testCases {
		gen := NewListGenerator()
		got, err := gen.GenerateParams(context.TODO(), &sourcev1.KustomizationSetGenerator{
			List: &sourcev1.ListGenerator{
				Elements: tt.elements,
			},
		}, nil)

		test.AssertNoError(t, err)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Fatalf("failed to generate pull requests:\n%s", diff)
		}
	}
}
