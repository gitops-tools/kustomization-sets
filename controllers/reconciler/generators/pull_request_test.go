package generators

import (
	"context"
	"testing"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/jenkins-x/go-scm/scm/factory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ Generator = (*PullRequestGenerator)(nil)

func TestGeneratePullRequestParams(t *testing.T) {
	testCases := []struct {
		dataFunc  func(*fakescm.Data)
		initObjs  []runtime.Object
		secretRef *corev1.LocalObjectReference
		labels    []string
		want      []map[string]string
	}{
		{
			dataFunc: func(d *fakescm.Data) {
				d.PullRequests[1] = &scm.PullRequest{
					Number: 1,
					Base: scm.PullRequestBranch{
						Ref: "main",
						Repo: scm.Repository{
							FullName: "test-org/my-repo",
						},
					},
					Head: scm.PullRequestBranch{
						Ref: "new-topic",
						Sha: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
					},
				}
			},
			want: []map[string]string{
				{
					"number":   "1",
					"branch":   "new-topic",
					"head_sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				},
			},
		},
	}

	for _, tt := range testCases {
		gen := NewPullRequestGenerator(logr.Discard(), fake.NewFakeClient(tt.initObjs...))
		client, data := fakescm.NewDefault()
		tt.dataFunc(data)
		gen.clientFactory = func(_, _, _ string, opts ...factory.ClientOptionFunc) (*scm.Client, error) {
			return client, nil
		}
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
