//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/civo/civogo"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetes) DeepCopyInto(out *CivoKubernetes) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetes.
func (in *CivoKubernetes) DeepCopy() *CivoKubernetes {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CivoKubernetes) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetesConnectionDetails) DeepCopyInto(out *CivoKubernetesConnectionDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetesConnectionDetails.
func (in *CivoKubernetesConnectionDetails) DeepCopy() *CivoKubernetesConnectionDetails {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetesConnectionDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetesList) DeepCopyInto(out *CivoKubernetesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CivoKubernetes, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetesList.
func (in *CivoKubernetesList) DeepCopy() *CivoKubernetesList {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CivoKubernetesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetesObservation) DeepCopyInto(out *CivoKubernetesObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetesObservation.
func (in *CivoKubernetesObservation) DeepCopy() *CivoKubernetesObservation {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetesObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetesParameters) DeepCopyInto(out *CivoKubernetesParameters) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetesParameters.
func (in *CivoKubernetesParameters) DeepCopy() *CivoKubernetesParameters {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetesParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetesSpec) DeepCopyInto(out *CivoKubernetesSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	if in.Pools != nil {
		in, out := &in.Pools, &out.Pools
		*out = make([]civogo.KubernetesClusterPoolConfig, len(*in))
		copy(*out, *in)
	}
	if in.Applications != nil {
		in, out := &in.Applications, &out.Applications
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.ConnectionDetails = in.ConnectionDetails
	if in.CNIPlugin != nil {
		in, out := &in.CNIPlugin, &out.CNIPlugin
		*out = new(string)
		**out = **in
	}
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetesSpec.
func (in *CivoKubernetesSpec) DeepCopy() *CivoKubernetesSpec {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoKubernetesStatus) DeepCopyInto(out *CivoKubernetesStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoKubernetesStatus.
func (in *CivoKubernetesStatus) DeepCopy() *CivoKubernetesStatus {
	if in == nil {
		return nil
	}
	out := new(CivoKubernetesStatus)
	in.DeepCopyInto(out)
	return out
}
