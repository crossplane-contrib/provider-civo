package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoSubnetConfig specs for the CivoSubnet
type CivoSubnetConfig struct {
	Name         string `json:"name"`
	NetworkID    string `json:"networkID"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceID"`
}

// CivoSubnetSpec holds the subnetConfig
type CivoSubnetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	SubnetConfig      CivoSubnetConfig `json:"subnetConfig"`
}

// CivoSubnetObservation observation fields
type CivoSubnetObservation struct {
	ID              string       `json:"id"`
	Status          string       `json:"status"`
	SubnetSize      string       `json:"subnetSize"`
	ObservableField string       `json:"observableField,omitempty"`
	CreatedAt       *metav1.Time `json:"createdAt,omitempty"`
}

// CivoSubnetStatus status of the resource
type CivoSubnetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoSubnetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CivoSubnet is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.atProvider.state"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoSubnetSpec   `json:"spec"`
	Status CivoSubnetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoSubnetList contains a list of CivoSubnet
type CivoSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoSubnet `json:"items"`
}
