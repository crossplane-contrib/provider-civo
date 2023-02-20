package civoclient

import (
	"github.com/civo/civogo"
)

const (
	// APIURL is the default Civo API URL
	APIURL = "https://api.civo.com"
)

// NewAPIClient returns a civogo client using the current default API key
func NewAPIClient(apiKey, region string) (*civogo.Client, error) {
	return civogo.NewClientWithURL(apiKey, APIURL, region)
}
