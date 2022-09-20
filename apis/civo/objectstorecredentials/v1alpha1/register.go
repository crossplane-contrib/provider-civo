package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "objectstorecredentials.civo.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// CivoObjectStoreCredentials type metadata.
var (
	CivoObjectStoreCredentialsKind             = reflect.TypeOf(CivoObjectStoreCredentials{}).Name()
	CivoObjectStoreCredentialsGroupKind        = schema.GroupKind{Group: Group, Kind: CivoObjectStoreCredentialsKind}.String()
	CivoObjectStoreCredentialsKindAPIVersion   = CivoObjectStoreCredentialsKind + "." + SchemeGroupVersion.String()
	CivoObjectStoreCredentialsGroupVersionKind = SchemeGroupVersion.WithKind(CivoObjectStoreCredentialsKind)
)

func init() {
	SchemeBuilder.Register(&CivoObjectStoreCredentials{}, &CivoObjectStoreCredentialsList{})
}
