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
// the ObjectMEta and KustomizationSpec.
type KustomizationSetTemplate struct {
	KustomizationSetTemplateMeta `json:"metadata"`
	Spec                         kustomizev1.KustomizationSpec `json:"spec"`
}

// ListGenerator include items info.
type ListGenerator struct {
	Elements []apiextensionsv1.JSON   `json:"elements"`
	Template KustomizationSetTemplate `json:"template,omitempty"`
}

// KustomizationSetGenerator include list item info
type KustomizationSetGenerator struct {
	List *ListGenerator `json:"list,omitempty"`
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
