package gitrepository

import (
	"context"
	"time"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/gitops-tools/kustomize-set-controller/pkg/reconciler/generators"
	"github.com/go-logr/logr"
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

func (g *GitRepositoryGenerator) Generate(_ context.Context, sg *sourcev1.KustomizationSetGenerator, ks *sourcev1.KustomizationSet) ([]map[string]any, error) {
	if sg == nil {
		return nil, generators.EmptyKustomizationSetGeneratorError
	}

	if sg.GitRepository == nil {
		return nil, nil
	}

	res := []map[string]any{}

	return res, nil
}

// Interval is an implementation of the Generator interface.
func (g *GitRepositoryGenerator) Interval(sg *sourcev1.KustomizationSetGenerator) time.Duration {
	return generators.NoRequeueInterval
}

// Template is an implementation of the Generator interface.
func (g *GitRepositoryGenerator) Template(sg *sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate {
	return sg.GitRepository.Template
}
