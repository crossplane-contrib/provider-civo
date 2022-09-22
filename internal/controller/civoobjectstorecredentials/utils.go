package civoobjectstorecredentials

import (
	"github.com/apex/log"
	"github.com/civo/civogo"
)

func FindObjectStoreCredentials(c *civogo.Client, search string) *civogo.ObjectStoreCredential {
	objectStoreCredentials, err := c.ListObjectStoreCredentials()
	if err != nil {
		log.Errorf("Unable to fetch object store %s", err)
		return nil
	}

	result := civogo.ObjectStoreCredential{}

	for _, value := range objectStoreCredentials.Items {
		if value.Name == search || value.ID == search {
			result = value
			return &result
		}
	}
	log.Infof("Object store was not found %s", search)
	return nil
}

func FindObjectStoreCredentialsCreds(c *civogo.Client, search string) *civogo.ObjectStoreCredential {
	objectstorecredentials, err := c.ListObjectStoreCredentials()
	if err != nil {
		log.Errorf("Unable to fetch object store credential %s", err)
		return nil
	}

	result := civogo.ObjectStoreCredential{}

	for _, value := range objectstorecredentials.Items {
		if value.AccessKeyID == search {
			result = value
			return &result
		}
	}
	log.Infof("Object store was not found %s", search)
	return nil
}
