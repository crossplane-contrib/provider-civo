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

// Package apis contains Kubernetes API for the Template provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	clusterv1alpha1 "github.com/crossplane-contrib/provider-civo/apis/civo/cluster/v1alpha1"
	instancev1alpha1 "github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
	objectstorev1alpha1 "github.com/crossplane-contrib/provider-civo/apis/civo/objectstore/v1alpha1"
	objectstorecredentialv1alpha1 "github.com/crossplane-contrib/provider-civo/apis/civo/objectstorecredential/v1alpha1"
	providerv1alpha1 "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		providerv1alpha1.SchemeBuilder.AddToScheme,
		clusterv1alpha1.SchemeBuilder.AddToScheme,
		instancev1alpha1.SchemeBuilder.AddToScheme,
		objectstorev1alpha1.SchemeBuilder.AddToScheme,
		objectstorecredentialv1alpha1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
