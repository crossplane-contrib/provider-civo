package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoObjectStoreCredentialObservation are the observable fields of a CivoObjectStoreCredentials.
type CivoObjectStoreCredentialObservation struct {
	ID        string `json:"id"`
	BucketURL string `json:"objectstoreCredentialsEndpoint"`
	Status    string `json:"status"`
}

// CivoObjectStoreCredentialConnectionDetails is the desired output secret to store connection information
type CivoObjectStoreCredentialConnectionDetails struct {
	ConnectionSecretNamePrefix string `json:"connectionSecretNamePrefix"`
	ConnectionSecretNamespace  string `json:"connectionSecretNamespace"`
}

// CivoObjectStoreCredentialSpec defines the spec for CivoObjectStoreCredentials
type CivoObjectStoreCredentialSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name of the CivoObjectStoreCredential
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Access key ID for the objects store credential
	// If not provided, one will be generated automatically
	// +optional
	AccessKeyID *string `json:"accessKeyID,omitempty"`

	// Location of the CivoObjectStoreCredentials Connection
	ConnectionDetails CivoObjectStoreCredentialConnectionDetails `json:"connectionDetails"`
}

// CivoObjectStoreCredentialStatus defines the status for CivoObjectStoreCredentials
type CivoObjectStoreCredentialStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoObjectStoreCredentialObservation `json:"atProvider,omitempty"`
	Message             string                               `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="cosc"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.atProvider.status",description="State of the Bucket"
// +kubebuilder:printcolumn:name="Bucket",type="string",JSONPath=".spec.name",description="Name of the Bucket which can be used against S3 API"

// CivoObjectStoreCredential is the Schema for the ObjectStoreCredentials API
type CivoObjectStoreCredential struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoObjectStoreCredentialSpec   `json:"spec,omitempty"`
	Status CivoObjectStoreCredentialStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoObjectStoreCredentialList contains a list of CivoObjectStoreCredentials
type CivoObjectStoreCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoObjectStoreCredential `json:"items"`
}
