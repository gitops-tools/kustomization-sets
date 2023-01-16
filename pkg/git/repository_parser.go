package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/fluxcd/pkg/http/fetch"
	"github.com/fluxcd/pkg/tar"
	kustomizationsetv1 "github.com/gitops-tools/kustomization-set-controller/api/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/yaml"
)

type archiveFetcher interface {
	Fetch(archiveURL, checksum, dir string) error
}

// retries is the number of retries to make when fetching artifacts.
const retries = 9

// RepositoryParser fetches archives from a GitRepository and parses the
// resources from them.
type RepositoryParser struct {
	fetcher archiveFetcher
	logr.Logger
}

// NewRepositoryParser creates and returns a RepositoryParser.
func NewRepositoryParser() *RepositoryParser {
	return &RepositoryParser{fetcher: fetch.NewArchiveFetcher(retries, tar.UnlimitedUntarSize, tar.UnlimitedUntarSize, "")}
}

// ParseFromArtifacts extracts the archive and processes the files.
func (p *RepositoryParser) ParseFromArtifacts(ctx context.Context, archiveURL, checksum string, dirs []kustomizationsetv1.GitRepositoryGeneratorDirectoryItem) ([]map[string]any, error) {
	tempDir, err := os.MkdirTemp("", "parsing")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory when parsing artifacts: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			p.Logger.Error(err, "failed to remove temporary archive directory")
		}
	}()

	if err := p.fetcher.Fetch(archiveURL, checksum, tempDir); err != nil {
		return nil, fmt.Errorf("failed to get archive URL %s: %w", archiveURL, err)
	}

	// TODO: exclude paths!

	result := []map[string]any{}
	for _, dir := range dirs {
		readPath, err := securejoin.SecureJoin(tempDir, dir.Path)
		if err != nil {
			return nil, err
		}
		files, err := os.ReadDir(readPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory from archive %q: %w", dir.Path, err)
		}

		for _, file := range files {
			// TODO: Limit this?
			localName := filepath.Join(dir.Path, file.Name())
			filename := filepath.Join(tempDir, localName)

			b, err := os.ReadFile(filename)
			if err != nil {
				return nil, fmt.Errorf("failed to read from archive file %s: %w", localName, err)
			}

			r := map[string]any{}
			if err := yaml.Unmarshal(b, &r); err != nil {
				return nil, fmt.Errorf("failed to parse archive file %s: %w", localName, err)
			}

			result = append(result, r)
		}
	}

	return result, nil
}
