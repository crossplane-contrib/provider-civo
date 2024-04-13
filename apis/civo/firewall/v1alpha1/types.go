package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CivoFirewallSpec defines the desired state of a Firewall.
type CivoFirewallSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Name is the name of the Firewall within Civo.
	// +kubebuilder:validation:Required
	// +immutable
	Name string `json:"name"`

	// NetworkID is the identifier for the network associated with the Firewall.
	// +kubebuilder:validation:Required
	// +immutable
	NetworkID string `json:"networkId"`

	// Region is the identifier for the region in which the Firewall is deployed.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// Rules are the set of rules applied to the firewall.
	// +optional
	Rules []FirewallRule `json:"rules,omitempty"`

	// ProviderReference holds configs (region, API key etc) for the crossplane provider that is being used.
	ProviderReference *xpv1.Reference `json:"providerReference"`
}

// FirewallRule defines the rules applied to the Firewall.
type FirewallRule struct {
	// Protocol used by the rule (TCP, UDP, ICMP).
	// +kubebuilder:validation:Enum=TCP;UDP;ICMP
	// +kubebuilder:validation:Required
	Protocol string `json:"protocol"`

	// StartPort is the starting port of the range.
	// +kubebuilder:validation:Required
	StartPort int `json:"startPort"`

	// EndPort is the ending port of the range.
	// +optional
	EndPort *int `json:"endPort,omitempty"`

	// CIDR is the IP address range that is applicable for the rule.
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr"`

	// Direction indicates whether the rule is for inbound or outbound traffic.
	// +kubebuilder:validation:Enum=ingress;egress
	// +kubebuilder:validation:Required
	Direction string `json:"direction"`

	// Label is an optional identifier for the rule.
	// +optional
	Label string `json:"label,omitempty"`
}

// CivoFirewallStatus defines the observed state of CivoFirewall.
type CivoFirewallStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoFirewallObservation `json:"atProvider,omitempty"`
}

// CivoFirewallObservation is used to reflect the observed state of the firewall.
type CivoFirewallObservation struct {
	// ID is the Civo ID of the Firewall.
	ID string `json:"id,omitempty"`

	// InstanceCount shows how many instances are using this firewall.
	InstanceCount *int `json:"instanceCount,omitempty"`

	// RulesCount shows how many rules are associated with this firewall.
	RulesCount int `json:"rulesCount"`
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

// CivoFirewallList contains a list of CivoFirewall.
type CivoFirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoFirewall `json:"items"`
}
