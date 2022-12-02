package git

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fluxcd/pkg/http/fetch"
	"github.com/fluxcd/pkg/tar"
	"gopkg.in/yaml.v3"
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
}

// NewRepositoryParser creates and returns a RepositoryParser.
func NewRepositoryParser() *RepositoryParser {
	return &RepositoryParser{fetcher: fetch.NewArchiveFetcher(retries, tar.UnlimitedUntarSize, tar.UnlimitedUntarSize, "")}
}

func (p *RepositoryParser) ParseFromArtifacts(ctx context.Context, archiveURL, checksum, parseDir string) ([]map[string]any, error) {
	tempDir, err := os.MkdirTemp("", "parsing")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory when parsing artifacts: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := p.fetcher.Fetch(archiveURL, checksum, tempDir); err != nil {
		return nil, fmt.Errorf("failed to get archive URL %s: %w", archiveURL, err)
	}

	files, err := os.ReadDir(filepath.Join(tempDir, parseDir))
	if err != nil {
		log.Fatal(err)
	}

	result := []map[string]any{}
	for _, file := range files {
		// TODO: Limit this?
		localName := filepath.Join(parseDir, file.Name())
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

	return result, nil
}
