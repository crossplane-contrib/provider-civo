package civoObjectStore

import (
	"fmt"

	"github.com/civo/civogo"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
)

func FindObjectStore(c *civogo.Client, search string) (*civogo.ObjectStore, error) {
	objectstores, err := c.ListObjectStores()
	if err != nil {
		return nil, err
	}

	result := civogo.ObjectStore{}

	for _, value := range objectstores.Items {
		if value.Name == search || value.ID == search {
			result = value
			return &result, nil
		}
	}
	err = fmt.Errorf("unable to find %s, zero matches", search)
	return nil, err

}

func FindObjectStoreViaKey(c *civogo.Client, search string) (*civogo.ObjectStore, error) {
	objectstores, err := c.ListObjectStores()
	if err != nil {
		return nil, err
	}

	result := civogo.ObjectStore{}

	for _, value := range objectstores.Items {
		if value.OwnerInfo.AccessKeyID == search {
			result = value
			return &result, nil
		}
	}
	err = fmt.Errorf("unable to find %s, zero matches", search)
	return nil, err

}

func connectionDetails(objectStore *civogo.ObjectStore, objectStoreCred *civogo.ObjectStoreCredential) (managed.ConnectionDetails, error) {
	if objectStore.Status == "ready" {
		return managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(objectStore.BucketURL),
			"accessKeyID":     []byte(objectStoreCred.AccessKeyID),
			"secretAccessKey": []byte(objectStoreCred.SecretAccessKeyID),
		}, nil
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(objectStore.BucketURL),
	}, nil
}
