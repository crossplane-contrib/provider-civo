/*
Copyright 2020 The Crossplane Authors.

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

// CivoObjectStoreObservation are the observable fields of a CivoObjectStore.
type CivoObjectStoreObservation struct {

	// User Given name to the Object Store
	Name string `json:"Name"`

	// Size of object store
	Size int64 `json:"Size"`

	// Status of the Object Store (e.g., Creating, Available, Deleting)
	Status string `json:"status,omitempty"`

	// Region where the Object Store is located
	Region string `json:"region,omitempty"`
}

// CivoObjectStoreConnectionDetails is the desired output secret to store connection information
type CivoObjectStoreConnectionDetails struct {
	ConnectionSecretNamePrefix string `json:"connectionSecretNamePrefix"`
	ConnectionSecretNamespace  string `json:"connectionSecretNamespace"`
}

// A CivoObjectStoreSpec defines the desired state of a CivoObjectStore.
type CivoObjectStoreSpec struct {
	// User Given name to the Object Store
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Size of object store, should be specified in GB
	// +kubebuilder:default:=500
	Size int64 `json:"size,omitempty"`

	ConnectionDetails CivoObjectStoreConnectionDetails `json:"connectionDetails"`
}

// A CivoObjectStoreStatus represents the observed state of a CivoObjectStore.
type CivoObjectStoreStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoObjectStoreObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CivoObjectStore is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.message"
// +kubebuilder:printcolumn:name="APPLICATIONS",type="string",JSONPath=".spec.applications"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoObjectStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoObjectStoreSpec   `json:"spec"`
	Status CivoObjectStoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoObjectStoreList contains a list of CivoObjectStore
type CivoObjectStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoObjectStore `json:"items"`
}
