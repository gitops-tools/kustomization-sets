package generators

import (
	"errors"
	"time"

	sourcev1 "github.com/gitops-tools/kustomize-set-controller/api/v1alpha1"
)

// Generator defines the interface implemented by all KustomizationSet generators.
type Generator interface {
	// GenerateParams interprets the KustomizationSet and generates all relevant parameters for the application template.
	// The expected / desired list of parameters is returned, it then will be render and reconciled
	// against the current state of the Applications in the cluster.
	GenerateParams(appSetGenerator *sourcev1.KustomizationSetGenerator, applicationSetInfo *sourcev1.KustomizationSet) ([]map[string]string, error)

	// GetRequeueAfter is the the generator can controller the next reconciled loop
	// In case there is more then one generator the time will be the minimum of the times.
	// In case NoRequeueAfter is empty, it will be ignored
	GetRequeueAfter(appSetGenerator *sourcev1.KustomizationSetGenerator) time.Duration

	// GetTemplate returns the inline template from the spec if there is any, or an empty object otherwise
	GetTemplate(appSetGenerator *sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate
}

// EmptyKustomizationSetGeneratorError is returned when KustomizationSet is empty.
var EmptyKustomizationSetGeneratorError = errors.New("KustomizationSet is empty")
var NoRequeueAfter time.Duration

// DefaultRequeueAfterSeconds is used when GetRequeueAfter is not specified, it is the default time to wait before the next reconcile loop
const DefaultRequeueAfterSeconds = 3 * time.Minute
