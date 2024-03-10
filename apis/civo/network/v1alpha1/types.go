package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoNetworkSpec holds the networkConfig
type CivoNetworkSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	Label             string `json:"label"`
}

// CivoNetworkObservation observation fields
type CivoNetworkObservation struct {
	ID              string       `json:"id"`
	CIDR            string       `json:"cidr"`
	Label           string       `json:"label"`
	Status          string       `json:"status"`
	Default         bool         `json:"default"`
	ObservableField string       `json:"observableField,omitempty"`
	CreatedAt       *metav1.Time `json:"createdAt,omitempty"`
}

// CivoNetworkStatus status of the resource
type CivoNetworkStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoNetworkObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CivoNetwork is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.atProvider.state"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoNetworkSpec   `json:"spec"`
	Status CivoNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoNetworkList contains a list of CivoNetwork
type CivoNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoNetwork `json:"items"`
}
