package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "objectStore.civo.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// CivoObjectStore type metadata.
var (
	CivoObjectStoreKind             = reflect.TypeOf(CivoObjectStore{}).Name()
	CivoObjectStoreGroupKind        = schema.GroupKind{Group: Group, Kind: CivoObjectStoreKind}.String()
	CivoObjectStoreKindAPIVersion   = CivoObjectStoreKind + "." + SchemeGroupVersion.String()
	CivoObjectStoreGroupVersionKind = SchemeGroupVersion.WithKind(CivoObjectStoreKind)
)

func init() {
	SchemeBuilder.Register(&CivoObjectStore{}, &CivoObjectStoreList{})
}
