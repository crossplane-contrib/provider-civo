package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoFirewallSpec holds the specs for firewall resource
type CivoFirewallSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name that you wish to use to refer to this firewall.
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// NetworkID for the network with which the firewall is to be associated.
	// +required
	// +immutable
	// +kubebuilder:validation:Required
	NetworkID string `json:"network_id"`

	// ProviderReference holds configs (region, API key etc) for the crossplane provider that is being used.
	ProviderReference *xpv1.Reference `json:"providerReference"`
}

// CivoFirewallObservation observation fields
type CivoFirewallObservation struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// CivoFirewallStatus status of the resource
type CivoFirewallStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoFirewallObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// CivoFirewall is the Schema for the CivoFirewalls API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.atProvider.state"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoFirewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoFirewallSpec   `json:"spec"`
	Status CivoFirewallStatus `json:"status,omitempty"`
}

// SetManagementPolicies sets up management policies.
func (mg *CivoFirewall) SetManagementPolicies(r xpv1.ManagementPolicies) {}

// GetManagementPolicies gets management policies.
func (mg *CivoFirewall) GetManagementPolicies() xpv1.ManagementPolicies {
	// Note: Crossplane runtime reconciler should leave handling of
	// ManagementPolicies to the provider controller. This is a temporary hack
	// until we remove the ManagementPolicy field from the Provider Kubernetes
	// Object in favor of the one in the ResourceSpec.
	return []xpv1.ManagementAction{xpv1.ManagementActionAll}
}

// SetPublishConnectionDetailsTo sets up connection details.
func (mg *CivoFirewall) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// GetPublishConnectionDetailsTo gets publish connection details.
func (mg *CivoFirewall) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// +kubebuilder:object:root=true

// CivoFirewallList contains a list of CivoFirewall
type CivoFirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoFirewall `json:"items"`
}
