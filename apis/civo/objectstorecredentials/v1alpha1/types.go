package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoObjectStoreCredentialsObservation are the observable fields of a CivoObjectStoreCredentials.
type CivoObjectStoreCredentialsObservation struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MaxSize   int    `json:"max_size"`
	BucketURL string `json:"ObjectStoreCredentials_endpoint"`
	Status    string `json:"status"`
}

// CivoObjectStoreCredentialsSpec defines the spec for CivoObjectStoreCredentials
type CivoObjectStoreCredentialsSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name is user-given name for the object store. It should be a S3 compatiable name
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +optional
	// Size should be specified in GB
	// +kubebuilder:default:=500
	MaxSizeGB int64 `json:"maxSize,omitempty"`

	// Name of the CivoObjectStoreCredentials access key
	// if the provided access key is found it'll be the owner
	// for object store, else a new credential will be created which can be accessed via the location given in connection details
	// +optional
	AccessKeyID       string  `json:"accessKeyId,omitempty"`
	SecretAccessKeyID *string `json:"secretAccessKeyID,omitempty"`

	// +optional
	// +kubebuilder:default:false
	// Suspended is a flag to indicate whether the credential is suspended.
	Suspended bool `json:"suspended"`
}

// CivoObjectStoreCredentialsStatus defines the status for CivoObjectStoreCredentials
type CivoObjectStoreCredentialsStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoObjectStoreCredentialsObservation `json:"atProvider,omitempty"`
	Message             string                                `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="cos"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.atProvider.status",description="State of the Bucket"
// +kubebuilder:printcolumn:name="Bucket",type="string",JSONPath=".spec.name",description="Name of the Bucket which can be used against S3 API"
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.maxSize",description="Size of the Bucket in GB"

// CivoObjectStoreCredentials is the Schema for the ObjectStoreCredentials API
type CivoObjectStoreCredentials struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoObjectStoreCredentialsSpec   `json:"spec,omitempty"`
	Status CivoObjectStoreCredentialsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoObjectStoreCredentialsList contains a list of CivoObjectStoreCredentials
type CivoObjectStoreCredentialsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoObjectStoreCredentials `json:"items"`
}
