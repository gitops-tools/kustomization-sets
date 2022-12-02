package list

import (
	"context"
	"reflect"
	"testing"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler/generators"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ generators.Generator = (*ListGenerator)(nil)

func TestGenerateListParams(t *testing.T) {
	testCases := []struct {
		name     string
		elements []apiextensionsv1.JSON
		want     []map[string]any
	}{
		{
			name:     "simple key/value pairs",
			elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url"}`)}},
			want:     []map[string]any{{"cluster": "cluster", "url": "url"}},
		},
		{
			name:     "nested key/values",
			elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url","values":{"foo":"bar"}}`)}},
			want:     []map[string]any{{"cluster": "cluster", "url": "url", "values": map[string]any{"foo": "bar"}}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			gen := NewGenerator()
			got, err := gen.Generate(context.TODO(), &sourcev1.KustomizationSetGenerator{
				List: &sourcev1.ListGenerator{
					Elements: tt.elements,
				},
			}, nil)

			test.AssertNoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("failed to generate list elements:\n%s", diff)
			}
		})
	}
}

func TestListGenerator_Interval(t *testing.T) {
	gen := NewGenerator()
	sg := &sourcev1.KustomizationSetGenerator{
		List: &sourcev1.ListGenerator{
			Elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url"}`)}},
		},
	}

	d := gen.Interval(sg)

	if d != generators.NoRequeueInterval {
		t.Fatalf("got %#v want %#v", d, generators.NoRequeueInterval)
	}
}

func TestListGenerator_GetTemplate(t *testing.T) {
	template := &sourcev1.KustomizationSetTemplate{
		KustomizationSetTemplateMeta: sourcev1.KustomizationSetTemplateMeta{
			Labels: map[string]string{
				"cluster.app/name": "{{ cluster }}",
			},
		},
	}
	gen := NewGenerator()
	sg := &sourcev1.KustomizationSetGenerator{
		List: &sourcev1.ListGenerator{
			Elements: []apiextensionsv1.JSON{{Raw: []byte(`{"cluster": "cluster","url": "url"}`)}},
			Template: template,
		},
	}

	tpl := gen.Template(sg)

	if !reflect.DeepEqual(tpl, template) {
		t.Fatalf("got %#v want %#v", tpl, template)
	}
}
