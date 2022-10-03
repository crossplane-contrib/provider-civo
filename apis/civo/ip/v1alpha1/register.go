package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "ip.civo.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// CivoIP type metadata.
var (
	CivoIPKind             = reflect.TypeOf(CivoIP{}).Name()
	CivoIPGroupKind        = schema.GroupKind{Group: Group, Kind: CivoIPKind}.String()
	CivoIPKindAPIVersion   = CivoIPKind + "." + SchemeGroupVersion.String()
	CivoIPGroupVersionKind = SchemeGroupVersion.WithKind(CivoIPKind)
)

func init() {
	SchemeBuilder.Register(&CivoIP{}, &CivoIPList{})
}
