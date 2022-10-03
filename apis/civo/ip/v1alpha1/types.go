package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoIPSpec holds the instanceConfig
type CivoIPSpec struct {
	xpv1.ResourceSpec `json:",inline"`
}

// CivoIPObservation observation fields
type CivoIPObservation struct {
	ID         string     `json:"id"`
	Address    string     `json:"address,omitempty"`
	AssignedTo AssignedTo `json:"assigned_to,omitempty"`
}

// AssignedTo represents IP assigned to resource
type AssignedTo struct {
	ID string `json:"id"`
	// Type can be one of the following:
	// - instance
	// - loadbalancer
	Type string `json:"type"`
	Name string `json:"name"`
}

// CivoIPStatus status of the resource
type CivoIPStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoIPObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CivoIP is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.atProvider.state"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoIPSpec   `json:"spec"`
	Status CivoIPStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CivoIPList contains a list of CivoIP
type CivoIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoIP `json:"items"`
}
