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

func (g *PullRequestGenerator) GenerateParams(ctx context.Context, sg *sourcev1.KustomizationSetGenerator, ks *sourcev1.KustomizationSet) ([]map[string]string, error) {
	if sg == nil {
		return nil, EmptyKustomizationSetGeneratorError
	}

	if sg.PullRequest == nil {
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

	scmClient, err := g.clientFactory(sg.PullRequest.Driver, sg.PullRequest.ServerURL, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	prs, _, err := scmClient.PullRequests.List(ctx, sg.PullRequest.Repo, listOptionsFromConfig(sg.PullRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	res := []map[string]string{}
	for _, pr := range prs {
		res = append(res, map[string]string{
			"number":   strconv.Itoa(pr.Number),
			"branch":   pr.Head.Ref,
			"head_sha": pr.Head.Sha,
		})
	}
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

// TODO: think about how to configure the limiting options.
func listOptionsFromConfig(c *sourcev1.PullRequestGenerator) scm.PullRequestListOptions {
	// TODO: labels!
	return scm.PullRequestListOptions{
		Size: 20,
	}
}

// // Determines which git-api protocol to use.
// // +kubebuilder:validation:Enum=github;gitlab;bitbucketserver
// // +optional
// Driver string `json:"driver"`
// // This is the API endpoint to use.
// // +kubebuilder:validation:Pattern="^https://"
// // +required
// ServerURL string `json:"serverURL"`
// // This should be the Repo you want to query.
// Repo string `json:"repo"`
// // The secret name containing the Git credentials.
// // For HTTPS repositories the secret must contain username and password
// // fields.
// // +optional
// SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`

// // Labels is used to filter the PRs that you want to target.
// Labels []string `json:"labels,omitempty"`
