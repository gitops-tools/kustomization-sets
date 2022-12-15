package gitrepository

import (
	"context"
	"fmt"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kustomizesetv1 "github.com/gitops-tools/kustomization-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomization-set-controller/pkg/generators"
	"github.com/gitops-tools/kustomization-set-controller/pkg/git"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GitRepositoryGenerator extracts files from Flux GitRepository resources.
type GitRepositoryGenerator struct {
	client.Client
	logr.Logger
}

// GeneratorFactory is a function for creating per-reconciliation generators the
// GitRepositoryGenerator.
func GeneratorFactory(c client.Client) generators.GeneratorFactory {
	return func(l logr.Logger) generators.Generator {
		return NewGenerator(l, c)
	}
}

// NewGenerator creates and returns a new GitRepository generator.
func NewGenerator(l logr.Logger, c client.Client) *GitRepositoryGenerator {
	return &GitRepositoryGenerator{
		Client: c,
		Logger: l,
	}
}

func (g *GitRepositoryGenerator) Generate(ctx context.Context, sg *kustomizesetv1.KustomizationSetGenerator, ks *kustomizesetv1.KustomizationSet) ([]map[string]any, error) {
	if sg == nil {
		return nil, generators.EmptyKustomizationSetGeneratorError
	}

	if sg.GitRepository == nil {
		return nil, nil
	}

	var gr sourcev1.GitRepository
	if err := g.Client.Get(ctx, client.ObjectKey{Name: sg.GitRepository.RepositoryRef, Namespace: ks.GetNamespace()}, &gr); err != nil {
		return nil, fmt.Errorf("could not load GitRepository: %w", err)
	}
	parser := git.NewRepositoryParser()

	return parser.ParseFromArtifacts(ctx, gr.Status.Artifact.URL, gr.Status.Artifact.Checksum, sg.GitRepository.Directories)
}

// Interval is an implementation of the Generator interface.
func (g *GitRepositoryGenerator) Interval(sg *kustomizesetv1.KustomizationSetGenerator) time.Duration {
	return generators.NoRequeueInterval
}

// Template is an implementation of the Generator interface.
func (g *GitRepositoryGenerator) Template(sg *kustomizesetv1.KustomizationSetGenerator) *kustomizesetv1.KustomizationSetTemplate {
	return sg.GitRepository.Template
}
