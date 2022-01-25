package generators

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clientFactoryFunc func(driver, serverURL, oauthToken string, opts ...factory.ClientOptionFunc) (*scm.Client, error)

// PullRequestGenerator generates from the open pull requests in a repository.
type PullRequestGenerator struct {
	client.Client
	clientFactory clientFactoryFunc
	logr.Logger
}

// NewPullRequestGenerator creates and returns a new pull request generator.
func NewPullRequestGenerator(l logr.Logger, c client.Client) *PullRequestGenerator {
	return &PullRequestGenerator{
		Client:        c,
		Logger:        l,
		clientFactory: factory.NewClient,
	}
}

func (g *PullRequestGenerator) Generate(ctx context.Context, sg *sourcev1.KustomizationSetGenerator, ks *sourcev1.KustomizationSet) ([]map[string]string, error) {
	g.Logger.Info("generating params", "repo", sg.PullRequest.Repo)
	if sg == nil {
		g.Logger.Info("no generator provided")
		return nil, EmptyKustomizationSetGeneratorError
	}

	if sg.PullRequest == nil {
		g.Logger.Info("pull request configuration is nil")
		return nil, nil
	}

	authToken := ""
	if sg.PullRequest.SecretRef != nil {
		secretName := types.NamespacedName{
			Namespace: ks.GetNamespace(),
			Name:      sg.PullRequest.SecretRef.Name,
		}

		var secret corev1.Secret
		if err := g.Get(ctx, secretName, &secret); err != nil {
			return nil, fmt.Errorf("failed to load repository generator credentials: %w", err)
		}
		// See https://github.com/fluxcd/source-controller/blob/main/pkg/git/options.go#L100
		// for details of the standard flux Git repository secret.
		authToken = string(secret.Data["password"])
	}
	g.Logger.Info("querying pull requests", "repo", sg.PullRequest.Repo, "driver", sg.PullRequest.Driver, "serverURL", sg.PullRequest.ServerURL)

	scmClient, err := g.clientFactory(sg.PullRequest.Driver, sg.PullRequest.ServerURL, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	prs, _, err := scmClient.PullRequests.List(ctx, sg.PullRequest.Repo, listOptionsFromConfig(sg.PullRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	g.Logger.Info("queried pull requests", "repo", sg.PullRequest.Repo, "count", len(prs))
	res := []map[string]string{}
	for _, pr := range prs {
		if !prMatchesLabels(pr, sg.PullRequest.Labels) {
			continue
		}
		res = append(res, map[string]string{
			"number":   strconv.Itoa(pr.Number),
			"branch":   pr.Head.Ref,
			"head_sha": pr.Head.Sha,
		})
	}
	return res, nil
}

// Interval is an implementation of the Generator interface.
func (g *PullRequestGenerator) Interval(sg *sourcev1.KustomizationSetGenerator) time.Duration {
	return sg.PullRequest.Interval.Duration
}

// Template is an implementation of the Generator interface.
func (g *PullRequestGenerator) Template(sg *sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate {
	return sg.PullRequest.Template
}

// label filtering is only supported by GitLab (that I'm aware of)
// The fetched PRs are filtered on labels across all providers, but providing
// the labels optimises the load from GitLab.
//
// TODO: How should we apply pagination/limiting of fetched PRs?
func listOptionsFromConfig(c *sourcev1.PullRequestGenerator) scm.PullRequestListOptions {
	return scm.PullRequestListOptions{
		Size:   20,
		Labels: c.Labels,
	}
}

func prMatchesLabels(pr *scm.PullRequest, labels []string) bool {
	if len(labels) == 0 {
		return true
	}
	for _, v := range pr.Labels {
		for _, l := range labels {
			if v.Name == l {
				return true
			}
		}
	}
	return false
}
