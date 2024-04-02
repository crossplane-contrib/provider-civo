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

// CivoVolumeObservation are the observable fields of a CivoVolume.
type CivoVolumeObservation struct {
	ID         string `json:"id"`
	InstanceID string `json:"instance_id"`
	Size       int    `json:"size"`
	Status     string `json:"status"`
}

// CivoVolumeSpec  defines schema for a CivoVolume resource.
type CivoVolumeSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name that you wish to use to refer to this volume.
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Size for the volume a minimum of 1 and a maximum of your available disk space from your quota specifies the size of the volume in gigabytes
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Size int `json:"Size"`

	// NetworkID for the network in which you wish to create the volume.
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	NetworkID string `json:"network_id"`

	// ClusterID is the identifier for the cluster to which this volume belongs, if applicable.
	// +optional
	ClusterID string `json:"cluster_id,omitempty"`

	// InstanceID is the identifier for the instance to which this volume is attached, if applicable.
	// +optional
	InstanceID string `json:"instance_id,omitempty"`

	// Bootable specifies whether the volume is bootable or not.
	// +optional
	Bootable bool `json:"bootable,omitempty"`

	// ProviderReference holds configs (region, API key etc.) for the crossplane provider that is being used.
	ProviderReference *xpv1.Reference `json:"providerReference"`
}

// A CivoVolumeStatus represents the observed state of a CivoVolume.
type CivoVolumeStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          *CivoVolumeObservation `json:"atProvider,omitempty"`
}

// SetManagementPolicies sets up management policies.
func (mg *CivoVolume) SetManagementPolicies(r xpv1.ManagementPolicies) {}

// GetManagementPolicies gets management policies.
func (mg *CivoVolume) GetManagementPolicies() xpv1.ManagementPolicies {
	// Note: Crossplane runtime reconciler should leave handling of
	// ManagementPolicies to the provider controller. This is a temporary hack
	// until we remove the ManagementPolicy field from the Provider Kubernetes
	// Object in favor of the one in the ResourceSpec.
	return []xpv1.ManagementAction{xpv1.ManagementActionAll}
}

// SetPublishConnectionDetailsTo sets up connection details.
func (mg *CivoVolume) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// GetPublishConnectionDetailsTo gets publish connection details.
func (mg *CivoVolume) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="cos"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.atProvider.status",description="State of the Bucket"
// +kubebuilder:printcolumn:name="Bucket",type="string",JSONPath=".spec.name",description="Name of the Bucket which can be used against S3 API"
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.maxSize",description="Size of the Bucket in GB"

// CivoVolume is the Schema for the CivoVolume API
type CivoVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoVolumeSpec   `json:"spec"`
	Status CivoVolumeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoVolumeList contains a list of CivoVolumeList
type CivoVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoVolume `json:"items"`
}
