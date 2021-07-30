package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "instance.civo.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// CivoInstance type metadata.
var (
	CivoInstanceKind            = reflect.TypeOf(CivoInstance{}).Name()
	CivoInstancGroupKind        = schema.GroupKind{Group: Group, Kind: CivoInstanceKind}.String()
	CivoInstancKindAPIVersion   = CivoInstanceKind + "." + SchemeGroupVersion.String()
	CivoInstancGroupVersionKind = SchemeGroupVersion.WithKind(CivoInstanceKind)
)

func init() {
	SchemeBuilder.Register(&CivoInstance{}, &CivoInstanceList{})
}
