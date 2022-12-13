package gitrepository

import (
	"context"
	"fmt"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	kustomizationsetv1 "github.com/gitops-tools/kustomization-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomization-set-controller/pkg/git"
	"github.com/gitops-tools/kustomization-set-controller/pkg/reconciler/generators"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GitRepositoryGenerator extracts files from Flux GitRepository resources.
type GitRepositoryGenerator struct {
	client.Client
	logr.Logger
}

// NewGenerator creates and returns a new pull request generator.
func NewGenerator(l logr.Logger, c client.Client) *GitRepositoryGenerator {
	return &GitRepositoryGenerator{
		Client: c,
		Logger: l,
	}
}

func (g *GitRepositoryGenerator) Generate(ctx context.Context, sg *kustomizationsetv1.KustomizationSetGenerator, ks *kustomizationsetv1.KustomizationSet) ([]map[string]any, error) {
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
func (g *GitRepositoryGenerator) Interval(sg *kustomizationsetv1.KustomizationSetGenerator) time.Duration {
	return generators.NoRequeueInterval
}

// Template is an implementation of the Generator interface.
func (g *GitRepositoryGenerator) Template(sg *kustomizationsetv1.KustomizationSetGenerator) *kustomizationsetv1.KustomizationSetTemplate {
	return sg.GitRepository.Template
}

// AdditionalResources is an implementation of the Generator interface.
func (g *GitRepositoryGenerator) AdditionalResources(*kustomizationsetv1.KustomizationSetGenerator) ([]runtime.Object, error) {
	return nil, nil
}
