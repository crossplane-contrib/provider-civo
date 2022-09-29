package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoObjectStoreObservation are the observable fields of a CivoObjectStore.
type CivoObjectStoreObservation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MaxSize   int    `json:"maxSize"`
	BucketURL string `json:"objectStoreEndpoint"`
	Status    string `json:"status"`
}

// CivoObjectStoreConnectionDetails is the desired output secret to store connection information
type CivoObjectStoreConnectionDetails struct {
	ConnectionSecretNamePrefix string `json:"connectionSecretNamePrefix"`
	ConnectionSecretNamespace  string `json:"connectionSecretNamespace"`
}

// CivoObjectStoreSpec defines the spec for CivoObjectStore
type CivoObjectStoreSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name is user-given name for the object store. It should be a S3 compatible name
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +optional
	// Size should be specified in GB
	// +kubebuilder:default:=500
	MaxSizeGB int64 `json:"maxSize,omitempty"`

	// Name of the CivoObjectStore access key
	// if the provided access key is found it'll be the owner for object store
	// +optional
	AccessKey string `json:"accessKey,omitempty"`

	// Location of the CivoObjectStore Connection
	ConnectionDetails CivoObjectStoreConnectionDetails `json:"connectionDetails"`
}

// CivoObjectStoreStatus defines the status for CivoObjectStore
type CivoObjectStoreStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoObjectStoreObservation `json:"atProvider,omitempty"`
	Message             string                     `json:"message"`
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

	Spec   CivoObjectStoreSpec   `json:"spec,omitempty"`
	Status CivoObjectStoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoObjectStoreList contains a list of CivoObjectStore
type CivoObjectStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoObjectStore `json:"items"`
}
