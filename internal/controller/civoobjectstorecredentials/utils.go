package civoobjectstorecredentials

import (
	"github.com/apex/log"
	"github.com/civo/civogo"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
)

// FindObjectStoreCredential finds objectstore credential
func FindObjectStoreCredential(c *civogo.Client, search string) *civogo.ObjectStoreCredential {
	objectStoreCredentials, err := c.ListObjectStoreCredentials()
	if err != nil {
		log.Errorf("Unable to fetch object store %s", err)
		return nil
	}

	result := civogo.ObjectStoreCredential{}
	for _, value := range objectStoreCredentials.Items {
		if value.Name == search || value.ID == search || value.AccessKeyID == search {
			result = value
			return &result
		}
	}
	log.Infof("Object store was not found %s", search)
	return nil
}

func connectionDetails(objectStoreCredential *civogo.ObjectStoreCredential) managed.ConnectionDetails {
	if objectStoreCredential.Status == "ready" {
		return managed.ConnectionDetails{
			"accessKeyID":     []byte(objectStoreCredential.AccessKeyID),
			"secretAccessKey": []byte(objectStoreCredential.SecretAccessKeyID),
		}
	}
	return nil
}
