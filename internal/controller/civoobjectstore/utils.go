package civoobjectstore

import (
	"github.com/apex/log"
	"github.com/civo/civogo"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
)

// FindObjectStore finds objectstore
func FindObjectStore(c *civogo.Client, search string) *civogo.ObjectStore {
	objectstores, err := c.ListObjectStores()
	if err != nil {
		log.Errorf("Unable to fetch object store %s", err)
		return nil
	}

	result := civogo.ObjectStore{}

	for _, value := range objectstores.Items {
		if value.Name == search || value.ID == search {
			result = value
			return &result
		}
	}
	log.Infof("Object store was not found %s", search)
	return nil
}

// FindObjectStoreCreds finds creds
func FindObjectStoreCreds(c *civogo.Client, search string) *civogo.ObjectStore {
	objectstores, err := c.ListObjectStores()
	if err != nil {
		log.Errorf("Unable to fetch object store credential %s", err)
		return nil
	}

	result := civogo.ObjectStore{}

	for _, value := range objectstores.Items {
		if value.OwnerInfo.AccessKeyID == search {
			result = value
			return &result
		}
	}
	log.Infof("Object store was not found %s", search)
	return nil
}

func connectionDetails(objectStore *civogo.ObjectStore, objectStoreCred *civogo.ObjectStoreCredential) managed.ConnectionDetails {
	if objectStore.Status == "ready" {
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
