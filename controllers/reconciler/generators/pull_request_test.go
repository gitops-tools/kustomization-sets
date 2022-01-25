package generators

import (
	"context"
	"reflect"
	"testing"
	"time"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/jenkins-x/go-scm/scm/factory"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ Generator = (*PullRequestGenerator)(nil)

func TestPullRequestGenerator_Generate(t *testing.T) {
	defaultClientFactory := func(c *scm.Client) clientFactoryFunc {
		return func(_, _, _ string, opts ...factory.ClientOptionFunc) (*scm.Client, error) {
			return c, nil
		}
	}

	testCases := []struct {
		name          string
		dataFunc      func(*fakescm.Data)
		initObjs      []runtime.Object
		secretRef     *corev1.LocalObjectReference
		labels        []string
		clientFactory func(*scm.Client) clientFactoryFunc
		want          []map[string]string
	}{
		{
			name: "simple unfiltered PR",
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
			clientFactory: defaultClientFactory,
			want: []map[string]string{
				{
					"number":   "1",
					"branch":   "new-topic",
					"head_sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				},
			},
		},
		{
			name: "filtering by label",
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
						Ref: "old-topic",
						Sha: "564254f7170844f40a01315fc571ae45fb8665b7",
					},
				}
				d.PullRequests[2] = &scm.PullRequest{
					Number: 2,
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
					Labels: []*scm.Label{{Name: "testing"}},
				}
			},
			labels:        []string{"testing"},
			clientFactory: defaultClientFactory,
			want: []map[string]string{
				{
					"number":   "2",
					"branch":   "new-topic",
					"head_sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewPullRequestGenerator(logr.Discard(), fake.NewFakeClient(tt.initObjs...))
			client, data := fakescm.NewDefault()
			tt.dataFunc(data)
			gen.clientFactory = tt.clientFactory(client)
			got, err := gen.Generate(context.TODO(), &sourcev1.KustomizationSetGenerator{
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
		})
	}
}

func TestPullRequestGenerator_GetInterval(t *testing.T) {
	interval := time.Minute * 10
	gen := NewPullRequestGenerator(logr.Discard(), fake.NewFakeClient())
	sg := &sourcev1.KustomizationSetGenerator{
		PullRequest: &sourcev1.PullRequestGenerator{
			Driver:    "fake",
			ServerURL: "https://example.com",
			Repo:      "test-org/my-repo",
			Interval:  metav1.Duration{Duration: interval},
		},
	}

	d := gen.Interval(sg)

	if d != interval {
		t.Fatalf("got %#v want %#v", d, interval)
	}
}

func TestPullRequestGenerator_GetTemplate(t *testing.T) {
	template := &sourcev1.KustomizationSetTemplate{
		KustomizationSetTemplateMeta: sourcev1.KustomizationSetTemplateMeta{
			Labels: map[string]string{
				"cluster.app/name": "{{ cluster }}",
			},
		},
	}
	gen := NewPullRequestGenerator(logr.Discard(), fake.NewFakeClient())
	sg := &sourcev1.KustomizationSetGenerator{
		PullRequest: &sourcev1.PullRequestGenerator{
			Driver:    "fake",
			ServerURL: "https://example.com",
			Repo:      "test-org/my-repo",
			Interval:  metav1.Duration{Duration: 10 * time.Minute},
			Template:  template,
		},
	}

	tpl := gen.Template(sg)

	if !reflect.DeepEqual(tpl, template) {
		t.Fatalf("got %#v want %#v", tpl, template)
	}
}
