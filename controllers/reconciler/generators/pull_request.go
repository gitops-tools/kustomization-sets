package generators

import (
	"context"
	"time"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PullRequestGenerator generates from the open pull requests in a repository.
type PullRequestGenerator struct {
	client.Client
}

// NewPullRequestGenerator creates and returns a new pull request generator.
func NewPullRequestGenerator(c client.Client) Generator {
	return &PullRequestGenerator{Client: c}
}

func (g *PullRequestGenerator) GenerateParams(ctx context.Context, sg *sourcev1.KustomizationSetGenerator, _ *sourcev1.KustomizationSet) ([]map[string]string, error) {
	if sg == nil {
		return nil, EmptyKustomizationSetGeneratorError
	}

	if sg.PullRequest == nil {
		return nil, nil
	}

	res := []map[string]string{}
	return res, nil
}

// GetInterval is an implementation of the Generator interface.
func (g *PullRequestGenerator) GetInterval(sg *sourcev1.KustomizationSetGenerator) time.Duration {
	return NoRequeueInterval
}

// GetTemplate is an implementation of the Generator interface.
func (g *PullRequestGenerator) GetTemplate(sg *sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate {
	return sg.List.Template
}
