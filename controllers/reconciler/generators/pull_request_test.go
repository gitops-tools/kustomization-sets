package generators

import (
	"context"
	"testing"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ Generator = (*PullRequestGenerator)(nil)

func TestGeneratePullRequestParams(t *testing.T) {
	testCases := []struct {
		initObjs  []runtime.Object
		secretRef *corev1.LocalObjectReference
		labels    []string
		want      []map[string]string
	}{
		{
			want: []map[string]string{
				{
					"number":   "1",
					"branch":   "main",
					"head_sha": "c31a9ba2da1d595b094491cc9ed67e551531984e",
				},
			},
		},
		{
			want: []map[string]string{},
		},
	}

	for _, tt := range testCases {
		gen := NewPullRequestGenerator(fake.NewFakeClient(tt.initObjs...))
		got, err := gen.GenerateParams(context.TODO(), &sourcev1.KustomizationSetGenerator{
			PullRequest: &sourcev1.PullRequestGenerator{
				Driver:    "fake",
				ServerURL: "https://example.com",
				Repo:      "test-org/my-repo",
				SecretRef: tt.secretRef,
				Labels:    tt.labels,
			},
		}, nil)

		test.AssertNoError(t, err)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Fatalf("failed to generate pull requests:\n%s", diff)
		}
	}
}
