/*
Copyright 2024 The Crossplane Authors.
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CivoNetworkObservation are the observable fields of a CivoNetwork.
type CivoNetworkObservation struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"default"`
	CIDR    string `json:"cidr"`

	// Used Capacity of the bucket in percentage.
	Label         string   `json:"label"`
	Status        string   `json:"status"`
	IPv4Enabled   bool     `json:"ipv4_enabled"`
	NameServersV4 []string `json:"nameservers_v4"`

	// Details regarding current state of the bucket.
	Conditions []metav1.Condition `json:"conditions"`
}

// CivoNetworkSpec  defines schema for a CivoNetwork resource.
type CivoNetworkSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name for the private network.
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// ProviderReference holds configs (region, API key etc.) for the crossplane provider that is being used.
	ProviderReference *xpv1.Reference `json:"providerReference"`
}

// A CivoNetworkStatus represents the observed state of a CivoNetwork.
type CivoNetworkStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          *CivoNetworkObservation `json:"atProvider,omitempty"`
}

// SetManagementPolicies sets up management policies.
func (mg *CivoNetwork) SetManagementPolicies(r xpv1.ManagementPolicies) {}

// GetManagementPolicies gets management policies.
func (mg *CivoNetwork) GetManagementPolicies() xpv1.ManagementPolicies {
	// Note: Crossplane runtime reconciler should leave handling of
	// ManagementPolicies to the provider controller. This is a temporary hack
	// until we remove the ManagementPolicy field from the Provider Kubernetes
	// Object in favor of the one in the ResourceSpec.
	return []xpv1.ManagementAction{xpv1.ManagementActionAll}
}

// SetPublishConnectionDetailsTo sets up connection details.
func (mg *CivoNetwork) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// GetPublishConnectionDetailsTo gets publish connection details.
func (mg *CivoNetwork) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="cos"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.atProvider.status",description="State of the Bucket"
// +kubebuilder:printcolumn:name="Bucket",type="string",JSONPath=".spec.name",description="Name of the Bucket which can be used against S3 API"
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.maxSize",description="Size of the Bucket in GB"

// CivoNetwork is the Schema for the CivoNetwork API
type CivoNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoNetworkSpec   `json:"spec"`
	Status CivoNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoNetworkList contains a list of CivoNetworkList
type CivoNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoNetwork `json:"items"`
}
