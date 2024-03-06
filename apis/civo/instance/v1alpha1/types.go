package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoInstanceConfig specs for the CivoInstance
type CivoInstanceConfig struct {
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// +immutable
	// +required
	Size string `json:"size,omitempty"`

	// +immutable
	// +required
	DiskImage string `json:"diskImage,omitempty"`

	// +optional
	Notes string `json:"notes,omitempty"`

	// +optional
	Script string `json:"script,omitempty"`

	// +required
	Region string `json:"region,omitempty"`

	// +optional
	Tags []string `json:"tags,omitempty"`

	// +immutable
	// +optional
	SSHPubKeyRef *SecretReference `json:"sshPubKeyRef,omitempty"`

	// +immutable
	// +optional
	InitialUser string `json:"initialUser,omitempty"`

	// +optional
	PublicIPRequired string `json:"publicIPRequired,omitempty"`
}

// SecretReference location of the SSH Public Key Secret
type SecretReference struct {
	// Name of the secret.
	Name string `json:"name"`

	// Namespace of the secret.
	Namespace string `json:"namespace"`

	// Key whose value will be used.
	Key string `json:"key"`
}

// CivoInstanceSpec holds the instanceConfig
type CivoInstanceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	InstanceConfig    CivoInstanceConfig `json:"instanceConfig"`
	// ProviderReference is the reference for the provider
	ProviderReference *xpv1.Reference `json:"providerReference,omitempty"`
}

// CivoInstanceObservation observation fields
type CivoInstanceObservation struct {
	ID              string       `json:"id"`
	State           string       `json:"state,omitempty"`
	IPv4            string       `json:"ipv4,omitempty"`
	ObservableField string       `json:"observableField,omitempty"`
	CreatedAt       *metav1.Time `json:"createdAt,omitempty"`
}

// CivoInstanceStatus status of the resource
type CivoInstanceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoInstanceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CivoInstance is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.atProvider.state"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoInstanceSpec   `json:"spec"`
	Status CivoInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoInstanceList contains a list of CivoInstance
type CivoInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoInstance `json:"items"`
}
