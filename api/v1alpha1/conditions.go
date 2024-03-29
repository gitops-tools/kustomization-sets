/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha1

import (
	"github.com/fluxcd/pkg/apis/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	maxConditionMessageLength = 20000

	// HealthyCondition indicates that the KustomizationSet has created all its
	// resources.
	HealthyCondition string = "Healthy"
)

// KustomizationSetReady registers a successful apply attempt of the given Kustomization.
func KustomizationSetReady(k KustomizationSet, inventory *ResourceInventory, reason, message string) KustomizationSet {
	setKustomizationSetReadiness(&k, metav1.ConditionTrue, reason, message)
	k.Status.Inventory = inventory
	return k
}

func setKustomizationSetReadiness(k *KustomizationSet, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    meta.ReadyCondition,
		Status:  status,
		Reason:  reason,
		Message: limitMessage(message),
	}
	apimeta.SetStatusCondition(&k.Status.Conditions, newCondition)
}

// chop a string and add an ellipsis to indicate that it's been chopped.
func limitMessage(s string) string {
	if len(s) <= maxConditionMessageLength {
		return s
	}

	return s[0:maxConditionMessageLength-3] + "..."
}

// ResourceInventory contains a list of Kubernetes resource object references that have been applied by a Kustomization.
type ResourceInventory struct {
	// Entries of Kubernetes resource object references.
	Entries []ResourceRef `json:"entries"`
}

// ResourceRef contains the information necessary to locate a resource within a cluster.
type ResourceRef struct {
	// ID is the string representation of the Kubernetes resource object's metadata,
	// in the format '<namespace>_<name>_<group>_<kind>'.
	ID string `json:"id"`

	// Version is the API version of the Kubernetes resource object's kind.
	Version string `json:"v"`
}
