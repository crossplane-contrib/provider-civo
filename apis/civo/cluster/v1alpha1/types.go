/*
Copyright 2020 The Crossplane Authors.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CivoKubernetesParameters are the configurable fields of a CivoKubernetes.
type CivoKubernetesParameters struct {
	ConfigurableField string `json:"configurableField"`
}

// CivoKubernetesObservation are the observable fields of a CivoKubernetes.
type CivoKubernetesObservation struct {
	ObservableField string `json:"observableField,omitempty"`
}

// CivoKubernetesConnectionDetails is the desired output secret to store connection information
type CivoKubernetesConnectionDetails struct {
	ConnectionSecretNamePrefix string `json:"connectionSecretNamePrefix"`
	ConnectionSecretNamespace  string `json:"connectionSecretNamespace"`
}

// A CivoKubernetesSpec defines the desired state of a CivoKubernetes.
type CivoKubernetesSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	Name              string                        `json:"name"`
	Pools             []KubernetesClusterPoolConfig `json:"pools"`
	// +optional
	// A list of applications to install from civo marketplace.
	Applications      []string                        `json:"applications,omitempty"`
	ConnectionDetails CivoKubernetesConnectionDetails `json:"connectionDetails"`
	// +optional
	// +kubebuilder:validation:Enum=flannel;cilium
	// +kubebuilder:default=flannel
	// +immutable
	// NOTE: This can only be set at creation time. Changing this value after creation will not update the CNI.
	CNIPlugin *string `json:"cni,omitempty"`
	// +optional
	// +kubebuilder:default="1.22.2-k3s1"
	// If not set, the default kubernetes version(1.22.2-k31) will be used.
	// If set, the value must be a valid kubernetes version, you can use the following command to get the valid versions: `civo k3s versions`
	// Changing the version to a higher version will upgrade the cluster. Note that this may cause breaking changes to the Kubernetes API so please check kubernetes deprecations/mitigations before upgrading.
	Version *string `json:"version,omitempty"`

	// ProviderReference holds configs (region, API key etc) for the crossplane provider that is being used.
	ProviderReference *xpv1.Reference `json:"providerReference"`
	// TODO: Update the examples as well
}

// A CivoKubernetesStatus represents the observed state of a CivoKubernetes.
type CivoKubernetesStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CivoKubernetesObservation `json:"atProvider,omitempty"`
	Message             string                    `json:"message"`
}

// +kubebuilder:object:root=true

// A CivoKubernetes is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.message"
// +kubebuilder:printcolumn:name="APPLICATIONS",type="string",JSONPath=".spec.applications"
// Please replace `PROVIDER-NAME` with your actual provider name, like `aws`, `azure`, `gcp`, `alibaba`
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,civo}
// +kubebuilder:subresource:status
type CivoKubernetes struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CivoKubernetesSpec   `json:"spec"`
	Status CivoKubernetesStatus `json:"status,omitempty"`
}

// SetManagementPolicies sets up management policies.
func (mg *CivoKubernetes) SetManagementPolicies(r xpv1.ManagementPolicies) {}

// GetManagementPolicies gets management policies.
func (mg *CivoKubernetes) GetManagementPolicies() xpv1.ManagementPolicies {
	// Note: Crossplane runtime reconciler should leave handling of
	// ManagementPolicies to the provider controller. This is a temporary hack
	// until we remove the ManagementPolicy field from the Provider Kubernetes
	// Object in favor of the one in the ResourceSpec.
	return []xpv1.ManagementAction{xpv1.ManagementActionAll}
}

// SetPublishConnectionDetailsTo sets up connection details.
func (mg *CivoKubernetes) SetPublishConnectionDetailsTo(r *xpv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// GetPublishConnectionDetailsTo gets publish connection details.
func (mg *CivoKubernetes) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// +kubebuilder:object:root=true

// CivoKubernetesList contains a list of CivoKubernetes
type CivoKubernetesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CivoKubernetes `json:"items"`
}

// KubernetesClusterPoolConfig defines the configuration for a pool of nodes in a Civo Kubernetes cluster.
// Should be converted to the equivalent type in the civogo package before being used.
// Should always be identical to the KubernetesClusterPoolConfig in the civogo package.
type KubernetesClusterPoolConfig struct {
	Region           string            `json:"region,omitempty"`
	ID               string            `json:"id,omitempty"`
	Count            int               `json:"count,omitempty"`
	Size             string            `json:"size,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	Taints           []corev1.Taint    `json:"taints"`
	PublicIPNodePool bool              `json:"public_ip_node_pool,omitempty"`
}
