//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoObjectStore) DeepCopyInto(out *CivoObjectStore) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoObjectStore.
func (in *CivoObjectStore) DeepCopy() *CivoObjectStore {
	if in == nil {
		return nil
	}
	out := new(CivoObjectStore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CivoObjectStore) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoObjectStoreConnectionDetails) DeepCopyInto(out *CivoObjectStoreConnectionDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoObjectStoreConnectionDetails.
func (in *CivoObjectStoreConnectionDetails) DeepCopy() *CivoObjectStoreConnectionDetails {
	if in == nil {
		return nil
	}
	out := new(CivoObjectStoreConnectionDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoObjectStoreList) DeepCopyInto(out *CivoObjectStoreList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CivoObjectStore, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoObjectStoreList.
func (in *CivoObjectStoreList) DeepCopy() *CivoObjectStoreList {
	if in == nil {
		return nil
	}
	out := new(CivoObjectStoreList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CivoObjectStoreList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoObjectStoreObservation) DeepCopyInto(out *CivoObjectStoreObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoObjectStoreObservation.
func (in *CivoObjectStoreObservation) DeepCopy() *CivoObjectStoreObservation {
	if in == nil {
		return nil
	}
	out := new(CivoObjectStoreObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoObjectStoreSpec) DeepCopyInto(out *CivoObjectStoreSpec) {
	*out = *in
	out.ConnectionDetails = in.ConnectionDetails
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoObjectStoreSpec.
func (in *CivoObjectStoreSpec) DeepCopy() *CivoObjectStoreSpec {
	if in == nil {
		return nil
	}
	out := new(CivoObjectStoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CivoObjectStoreStatus) DeepCopyInto(out *CivoObjectStoreStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CivoObjectStoreStatus.
func (in *CivoObjectStoreStatus) DeepCopy() *CivoObjectStoreStatus {
	if in == nil {
		return nil
	}
	out := new(CivoObjectStoreStatus)
	in.DeepCopyInto(out)
	return out
}
