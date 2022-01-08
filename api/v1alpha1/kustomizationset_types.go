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
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KustomizationSetTemplateMeta represents the metadata  fields that may
// be used for Kustomizations generated from the KustomizationSet (based on metav1.ObjectMeta)
type KustomizationSetTemplateMeta struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Finalizers  []string          `json:"finalizers,omitempty"`
}

// KustomizationSetTemplate represents Kustomization specs as a split between
// the ObjectMeta and KustomizationSpec.
type KustomizationSetTemplate struct {
	KustomizationSetTemplateMeta `json:"metadata"`
	Spec                         kustomizev1.KustomizationSpec `json:"spec"`
}

// ListGenerator generates from a hard-coded list.
type ListGenerator struct {
	Elements []apiextensionsv1.JSON    `json:"elements"`
	Template *KustomizationSetTemplate `json:"template,omitempty"`
}

// GitDirectoryGeneratorItem represents a path to be scanned in a repository.
type GitDirectoryGeneratorItem struct {
	Path    string `json:"path"`
	Exclude bool   `json:"exclude,omitempty"`
}

// GitRepositoryGenerator generates from a Flux GitRepository.
type GitRepositoryGenerator struct {
	RepositoryRef corev1.LocalObjectReference `json:"repositoryRef"`
	Directories   []GitDirectoryGeneratorItem `json:"directories,omitempty"`

	// The interval at which to check for repository updates.
	// +required
	Interval metav1.Duration           `json:"interval"`
	Template *KustomizationSetTemplate `json:"template,omitempty"`
}

// PullRequestGenerator defines a generator that queries a Git hosting service
// for relevant PRs.
type PullRequestGenerator struct {
	// The interval at which to check for repository updates.
	// +required
	Interval metav1.Duration           `json:"interval"`
	Template *KustomizationSetTemplate `json:"template,omitempty"`

	// TODO: Fill this out with the rest of the elements from
	// https://github.com/jenkins-x/go-scm/blob/main/scm/factory/factory.go

	// Determines which git-api protocol to use.
	// +kubebuilder:validation:Enum=github;gitlab;bitbucketserver
	// +optional
	Driver string `json:"driver"`
	// This is the API endpoint to use.
	// +kubebuilder:validation:Pattern="^https://"
	ServerURL string `json:"serverURL,omitempty"`
	// This should be the Repo you want to query.
	Repo string `json:"repo"`
	// The secret name containing the Git credentials.
	// For HTTPS repositories the secret must contain username and password
	// fields.
	// +optional
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`

	// Labels is used to filter the PRs that you want to target.
	Labels []string `json:"labels,omitempty"`
}

// KustomizationSetGenerator include list item info
type KustomizationSetGenerator struct {
	List          *ListGenerator          `json:"list,omitempty"`
	GitRepository *GitRepositoryGenerator `json:"gitRepository,omitempty"`
	PullRequest   *PullRequestGenerator   `json:"pullRequest,omitempty"`
}

// KustomizationSetSpec defines the desired state of KustomizationSet
type KustomizationSetSpec struct {
	Generators []KustomizationSetGenerator `json:"generators"`
	Template   KustomizationSetTemplate    `json:"template"`
}

// KustomizationSetStatus defines the observed state of KustomizationSet
type KustomizationSetStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KustomizationSet is the Schema for the kustomizationsets API
type KustomizationSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KustomizationSetSpec   `json:"spec,omitempty"`
	Status KustomizationSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KustomizationSetList contains a list of KustomizationSet
type KustomizationSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KustomizationSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KustomizationSet{}, &KustomizationSetList{})
}
