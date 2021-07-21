package civocli

import (
	"strings"

	"github.com/civo/civogo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	defaultTags = ""
)

// CivoClient is a client for communicating with Civo.
type CivoClient struct {
	apikey       string
	civoGoClient *civogo.Client
}

// NewCivoClient creates a new Civo client.
func NewCivoClient(apiKey string, region string) (*CivoClient, error) {

	if apiKey == "" {
		return nil, errors.New("newCivoClient: apiKey is nil")
	}

	apiKey = strings.TrimSuffix(apiKey, "\n")

	if region == "" {
		return nil, errors.New("newCivoClient: region is nil")
	}
	client, err := civogo.NewClient(apiKey, region)
	if err != nil {
		return nil, err
	}
	return &CivoClient{
		apikey:       apiKey,
		civoGoClient: client,
	}, nil
}

// GetK3sCluster gets a K3s cluster on Civo.
func (c *CivoClient) GetK3sCluster(clusterName string) (*civogo.KubernetesCluster, error) {

	kubernetesCluster, err := c.civoGoClient.FindKubernetesCluster(clusterName)
	if err != nil {
		if strings.Contains(err.Error(), "ZeroMatchesError") {
			return nil, nil
		}
		return nil, err
	}
	return kubernetesCluster, nil
}

// CreateNewK3sCluster creates a new K3s cluster on Civo.
func (c *CivoClient) CreateNewK3sCluster(clusterName string, numberOfInstances int, instanceSize string, applications []string) error {

	// Find the default network ID
	network, err := c.civoGoClient.GetDefaultNetwork()
	if err != nil {
		return err
	}

	cfg := &civogo.KubernetesClusterConfig{
		Name:            clusterName,
		Tags:            defaultTags,
		NetworkID:       network.ID,
		NumTargetNodes:  numberOfInstances,
		TargetNodesSize: instanceSize,
		Applications:    strings.Join(applications, ","),
	}

	kubernetesCluster, err := c.civoGoClient.NewKubernetesClusters(cfg)

	if err != nil {
		return err
	}

	log.Debugf("Created Kubernetes cluster %s with %d instances", kubernetesCluster.Name, len(kubernetesCluster.Instances))

	return nil
}

// DeleteK3sCluster deletes a k3s cluster on Civo.
func (c *CivoClient) DeleteK3sCluster(name string) error {

	kubernetesCluster, err := c.GetK3sCluster(name)
	if err != nil {
		return err
	}
	resp, err := c.civoGoClient.DeleteKubernetesCluster(kubernetesCluster.ID)
	if err != nil {
		log.Debugf("error [%s %s %s %s]", resp.Result, resp.ErrorDetails, resp.ErrorCode, resp.ErrorReason)
	}
	return err
}
