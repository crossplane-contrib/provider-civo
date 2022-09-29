package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "objectstorecredential.civo.crossplane.io"
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
	CivoObjectStoreCredentialKind             = reflect.TypeOf(CivoObjectStoreCredential{}).Name()
	CivoObjectStoreCredentialGroupKind        = schema.GroupKind{Group: Group, Kind: CivoObjectStoreCredentialKind}.String()
	CivoObjectStoreCredentialKindAPIVersion   = CivoObjectStoreCredentialKind + "." + SchemeGroupVersion.String()
	CivoObjectStoreCredentialGroupVersionKind = SchemeGroupVersion.WithKind(CivoObjectStoreCredentialKind)
)

func init() {
	SchemeBuilder.Register(&CivoObjectStoreCredential{}, &CivoObjectStoreCredentialList{})
}
