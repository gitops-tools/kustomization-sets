package git

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"

	kustomizationsetv1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/test"
	"github.com/google/go-cmp/cmp"
)

func TestFetchArchiveResources(t *testing.T) {
	fetchTests := []struct {
		description string
		filename    string
		want        []map[string]any
	}{
		{
			description: "simple yaml files",
			filename:    "/files.tar.gz",
			want: []map[string]any{
				{"environment": "dev", "instances": 2.0},
				{"environment": "production", "instances": 10.0},
				{"environment": "staging", "instances": 5.0},
			},
		},
		{
			description: "simple json files",
			filename:    "/json_files.tar.gz",
			want: []map[string]any{
				{"environment": "dev", "instances": 1.0},
				{"environment": "production", "instances": 10.0},
				{"environment": "staging", "instances": 5.0},
			},
		},
	}

	srv := test.StartFakeArchiveServer(t, "testdata")
	for _, tt := range fetchTests {
		t.Run(tt.description, func(t *testing.T) {
			parser := NewRepositoryParser()
			parsed, err := parser.ParseFromArtifacts(context.TODO(), srv.URL+tt.filename, strings.TrimSpace(mustReadFile(t, "testdata"+tt.filename+".sum")), []kustomizationsetv1.GitRepositoryGeneratorDirectoryItem{{Path: "files"}})
			if err != nil {
				t.Fatal(err)
			}
			sort.Slice(parsed, func(i, j int) bool { return parsed[i]["environment"].(string) < parsed[j]["environment"].(string) })
			if diff := cmp.Diff(tt.want, parsed); diff != "" {
				t.Fatalf("failed to parse artifacts:\n%s", diff)
			}
		})
	}
}

func TestFetchArchiveResources_bad_yaml(t *testing.T) {
	parser := NewRepositoryParser()
	srv := test.StartFakeArchiveServer(t, "testdata")

	_, err := parser.ParseFromArtifacts(context.TODO(), srv.URL+"/bad_files.tar.gz", strings.TrimSpace(mustReadFile(t, "testdata/bad_files.tar.gz.sum")), []kustomizationsetv1.GitRepositoryGeneratorDirectoryItem{{Path: "files"}})
	if err.Error() != `failed to parse archive file files/dev.yaml: error converting YAML to JSON: yaml: line 4: could not find expected ':'` {
		t.Fatalf("got error %v", err)
	}
}

func mustReadFile(t *testing.T, filename string) string {
	t.Helper()
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}
