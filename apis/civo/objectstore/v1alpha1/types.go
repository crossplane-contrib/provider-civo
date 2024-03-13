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

// CivoObjectStoreObservation are the observable fields of a CivoObjectStore.
type CivoObjectStoreObservation struct {

	// Total size of object store.
	MaxSize int64 `json:"maxSize"`

	// Used Capacity of the bucket.
	UsedCapacity string `json:"usedCapacity"`

	// Status of the Object Store.
	Status CivoObjectStoreStatus `json:"status,omitempty"`

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
	xpv1.ResourceSpec `json:",inline"`

	// Name for object store(aka bucket). Name  needs to be unique across the entire region
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Total size of object store, should be specified in GB and should be multiple of 500
	// +kubebuilder:default:=500
	MaxSize int `json:"maxSize,omitempty"`

	// Name of the CivoObjectStore access key
	// if the provided access key is found it'll be the owner
	// for object store, else a new credential will be created which can be accessed via the location given in connection details
	// +optional
	AccessKey string `json:"accessKey,omitempty"`

	// ConnectionDetails specifies how the connection information for the object store should be stored.
	// This includes the necessary details to create a secret that can store sensitive information
	// such as the access key and secret access key securely.
	ConnectionDetails CivoObjectStoreConnectionDetails `json:"connectionDetails"`
}

// A CivoObjectStoreStatus represents the observed state of a CivoObjectStore.
type CivoObjectStoreStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          *CivoObjectStoreObservation `json:"atProvider,omitempty"`
	Conditions          []metav1.Condition          `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="cos"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.atProvider.status",description="State of the Bucket"
// +kubebuilder:printcolumn:name="Bucket",type="string",JSONPath=".spec.name",description="Name of the Bucket which can be used against S3 API"
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.maxSize",description="Size of the Bucket in GB"

// CivoObjectStore is the Schema for the ObjectStore API
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
