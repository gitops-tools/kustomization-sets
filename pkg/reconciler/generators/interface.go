package generators

import (
	"context"
	"errors"
	"time"

	sourcev1 "github.com/gitops-tools/kustomization-set-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Generator defines the interface implemented by all KustomizationSet generators.
type Generator interface {
	// Generate interprets the KustomizationSet and generates all relevant
	// parameters for the Kustomization template.
	// The expected / desired list of parameters is returned, it then will be render and reconciled
	// against the current state of the Applications in the cluster.
	Generate(context.Context, *sourcev1.KustomizationSetGenerator, *sourcev1.KustomizationSet) ([]map[string]any, error)

	// Interval is the the generator can controller the next reconciled loop
	//
	// In case there is more then one generator the time will be the minimum of the times.
	// In case NoRequeueInterval is empty, it will be ignored
	Interval(*sourcev1.KustomizationSetGenerator) time.Duration

	// Template returns the inline template from the spec if there is any, or
	// an empty object otherwise
	Template(*sourcev1.KustomizationSetGenerator) *sourcev1.KustomizationSetTemplate

	// AdditionalResources returns a set of additional templatable resources
	// that this generator needs to have created to be used with the
	// Kustomization.
	//
	// Generators can return nil to indicate no additional resources.
	AdditionalResources(*sourcev1.KustomizationSetGenerator) ([]runtime.Object, error)
}

// EmptyKustomizationSetGeneratorError is returned when KustomizationSet is
// empty.
var EmptyKustomizationSetGeneratorError = errors.New("KustomizationSet is empty")
var NoRequeueInterval time.Duration

// DefaultInterval is used when Interval is not specified, it
// is the default time to wait before the next reconcile loop.
const DefaultRequeueAfterSeconds = 3 * time.Minute
