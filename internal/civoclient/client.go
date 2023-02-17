package civoclient

import (
	"github.com/civo/civogo"
)

const (
	APIURL = "https://api.civo.com"
)

// CivoAPIClient returns a civogo client using the current default API key
func NewAPIClient(APIKey, region string) (*civogo.Client, error) {
	return civogo.NewClientWithURL(APIKey, APIURL, region)
}
