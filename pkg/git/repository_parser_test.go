package git

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

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
				{"environment": "dev", "instances": 2},
				{"environment": "production", "instances": 10},
				{"environment": "staging", "instances": 5},
			},
		},
		{
			description: "simple json files",
			filename:    "/json_files.tar.gz",
			want: []map[string]any{
				{"environment": "dev", "instances": 1},
				{"environment": "production", "instances": 10},
				{"environment": "staging", "instances": 5},
			},
		},
	}

	srv := startFakeArchiveServer(t)
	for _, tt := range fetchTests {
		t.Run(tt.description, func(t *testing.T) {
			parser := NewRepositoryParser()
			parsed, err := parser.ParseFromArtifacts(context.TODO(), srv.URL+tt.filename, strings.TrimSpace(mustReadFile(t, "testdata"+tt.filename+".sum")), "files")
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
	srv := startFakeArchiveServer(t)

	_, err := parser.ParseFromArtifacts(context.TODO(), srv.URL+"/bad_files.tar.gz", strings.TrimSpace(mustReadFile(t, "testdata/bad_files.tar.gz.sum")), "files")
	if err.Error() != `failed to parse archive file files/dev.yaml: yaml: line 3: could not find expected ':'` {
		t.Fatalf("got error %v", err)
	}
}

func startFakeArchiveServer(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	t.Cleanup(ts.Close)

	return ts
}

func mustReadFile(t *testing.T, filename string) string {
	t.Helper()
	b, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}
