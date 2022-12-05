package gitrepository

import (
	"context"
	"reflect"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kustomizesetv1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler/generators"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ generators.Generator = (*GitRepositoryGenerator)(nil)

const testNamespace = "generation"

func TestGitRepositoryGenerator_Params(t *testing.T) {
	srv := test.StartFakeArchiveServer(t, "testdata")
	testCases := []struct {
		name      string
		generator *kustomizesetv1.GitRepositoryGenerator
		objects   []runtime.Object
		want      []map[string]any
	}{
		{
			"simple case",
			&kustomizesetv1.GitRepositoryGenerator{
				RepositoryRef: "test-repository",
				Directories: []kustomizesetv1.GitRepositoryGeneratorDirectoryItem{
					{Path: "files"},
				},
			},
			[]runtime.Object{newGitRepository(srv.URL+"/files.tar.gz",
				"f0a57ec1cdebda91cf00d89dfa298c6ac27791e7fdb0329990478061755eaca8")},
			[]map[string]any{
				{"environment": "dev", "instances": 2.0},
				{"environment": "production", "instances": 10.0},
				{"environment": "staging", "instances": 5.0},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(logr.Discard(), newFakeClient(t, tt.objects...))
			got, err := gen.Generate(context.TODO(), &kustomizesetv1.KustomizationSetGenerator{
				GitRepository: tt.generator,
			},
				&kustomizesetv1.KustomizationSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-generator",
						Namespace: testNamespace,
					},
					Spec: kustomizesetv1.KustomizationSetSpec{
						Generators: []kustomizesetv1.KustomizationSetGenerator{
							{
								GitRepository: tt.generator,
							},
						},
					},
				})

			test.AssertNoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("failed to generate pull requests:\n%s", diff)
			}
		})
	}
}

func TestGitRepositoryGenerator_Interval(t *testing.T) {
	gen := NewGenerator(logr.Discard(), nil)
	sg := &kustomizesetv1.KustomizationSetGenerator{
		GitRepository: &kustomizesetv1.GitRepositoryGenerator{},
	}

	d := gen.Interval(sg)

	if d != generators.NoRequeueInterval {
		t.Fatalf("got %#v want %#v", d, generators.NoRequeueInterval)
	}
}

func TestGitRepositoryGenerator_GetTemplate(t *testing.T) {
	template := &kustomizesetv1.KustomizationSetTemplate{
		KustomizationSetTemplateMeta: kustomizesetv1.KustomizationSetTemplateMeta{
			Labels: map[string]string{
				"cluster.app/name": "{{ cluster }}",
			},
		},
	}
	gen := NewGenerator(logr.Discard(), nil)
	sg := &kustomizesetv1.KustomizationSetGenerator{
		GitRepository: &kustomizesetv1.GitRepositoryGenerator{
			Template: template,
		},
	}

	tpl := gen.Template(sg)

	if !reflect.DeepEqual(tpl, template) {
		t.Fatalf("got %#v want %#v", tpl, template)
	}
}

func newGitRepository(archiveURL, xsum string) *sourcev1.GitRepository {
	return &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repository",
			Namespace: testNamespace,
		},
		Status: sourcev1.GitRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				URL:      archiveURL,
				Checksum: xsum,
			},
		},
	}
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.WithWatch {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := sourcev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := kustomizesetv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()
}
