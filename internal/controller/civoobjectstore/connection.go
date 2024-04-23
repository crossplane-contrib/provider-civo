package civoobjectstore

import (
	"github.com/civo/civogo"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
)

func connectionDetails(objectStore *civogo.ObjectStore, objectStoreCred *civogo.ObjectStoreCredential) managed.ConnectionDetails {
	if objectStore.Status == objectStoreStatusReady {
		return managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(objectStore.BucketURL),
			"accessKeyID":     []byte(objectStoreCred.AccessKeyID),
			"secretAccessKey": []byte(objectStoreCred.SecretAccessKeyID),
		}
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(objectStore.BucketURL),
	}
}
