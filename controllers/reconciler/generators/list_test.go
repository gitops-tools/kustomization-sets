package generators

import (
	"testing"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"
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

	for _, testCase := range testCases {
		listGenerator := NewListGenerator()
		got, err := listGenerator.GenerateParams(&sourcev1.KustomizationSetGenerator{
			List: &sourcev1.ListGenerator{
				Elements: testCase.elements,
			},
		}, nil)

		assert.NoError(t, err)
		assert.ElementsMatch(t, testCase.want, got)
	}
}
