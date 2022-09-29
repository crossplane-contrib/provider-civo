package civocli

import (
	"strings"

	"github.com/civo/civogo"
	providerCivoCluster "github.com/crossplane-contrib/provider-civo/apis/civo/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	defaultTags = ""
)

const (
	// StateActive instance is ready to use
	StateActive = "ACTIVE"
	// StateBuilding instance is still building
	StateBuilding = "BUILDING"
)

// CivoClient is a client for communicating with Civo.
type CivoClient struct {
	apikey       string
	CivoGoClient *civogo.Client
}

func emptyIfNil(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

// GenerateObservation creates the CivoInstanceObservation from instance infos
func GenerateObservation(instance *civogo.Instance) (v1alpha1.CivoInstanceObservation, error) {
	observation := v1alpha1.CivoInstanceObservation{
		ID:    instance.ID,
		State: instance.Status,
		IPv4:  instance.PublicIP,
	}

	if !observation.CreatedAt.IsZero() {
		if err := observation.CreatedAt.UnmarshalText([]byte(instance.CreatedAt.String())); err != nil {
			return v1alpha1.CivoInstanceObservation{}, errors.Wrap(err, "errUnmarshalDate")
		}
	}
	return observation, nil
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
		CivoGoClient: client,
	}, nil
}

// UpdateInstance updates a civo instance
func (c *CivoClient) UpdateInstance(id string, instance *v1alpha1.CivoInstance) error {
	civoInstance, err := c.CivoGoClient.GetInstance(id)
	if err != nil {
		return err
	}
	if civoInstance.Hostname != instance.Spec.InstanceConfig.Hostname {
		civoInstance.Hostname = instance.Spec.InstanceConfig.Hostname
	}
	if civoInstance.Notes != instance.Spec.InstanceConfig.Notes {
		civoInstance.Notes = instance.Spec.InstanceConfig.Notes
	}
	_, err = c.CivoGoClient.UpdateInstance(civoInstance)
	if err != nil {
		return err
	}
	return nil
}

// CreateNewInstance creates a new instance on Civo.
func (c *CivoClient) CreateNewInstance(instance *v1alpha1.CivoInstance, sshPubKey, diskImageName string) (*civogo.Instance, error) {
	config, err := c.CivoGoClient.NewInstanceConfig()
	if err != nil {
		return nil, err
	}
	config.Hostname = emptyIfNil(&instance.Spec.InstanceConfig.Hostname)
	config.Size = instance.Spec.InstanceConfig.Size
	config.Tags = instance.Spec.InstanceConfig.Tags
	config.Script = emptyIfNil(&instance.Spec.InstanceConfig.Script)
	config.Region = instance.Spec.InstanceConfig.Region
	config.InitialUser = emptyIfNil(&instance.Spec.InstanceConfig.InitialUser)
	config.PublicIPRequired = emptyIfNil(&instance.Spec.InstanceConfig.PublicIPRequired)

	if len(sshPubKey) > 0 {
		if sshKey, err := c.CivoGoClient.FindSSHKey(config.Hostname); err == nil {
			config.SSHKeyID = sshKey.ID
		} else {
			newSSHKey, err := c.CivoGoClient.NewSSHKey(config.Hostname, sshPubKey)
			if err != nil {
				return nil, err
			}
			config.SSHKeyID = newSSHKey.ID
		}
	}

	template, err := c.CivoGoClient.FindDiskImage(diskImageName)
	if err != nil {
		return nil, err
	}
	config.TemplateID = template.ID
	result, err := c.CivoGoClient.CreateInstance(config)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteInstance deletes a instance on Civo.
func (c *CivoClient) DeleteInstance(id string) error {
	instance, err := c.CivoGoClient.GetInstance(id)
	if err != nil {
		return err
	}
	resp, err := c.CivoGoClient.DeleteInstance(instance.ID)
	if err != nil {
		log.Debugf("error [%s %s %s %s]", resp.Result, resp.ErrorDetails, resp.ErrorCode, resp.ErrorReason)
	}
	if sshKey, err := c.CivoGoClient.FindSSHKey(instance.Hostname); err == nil {
		_, err := c.CivoGoClient.DeleteSSHKey(sshKey.ID)
		if err != nil {
			return err
		}
	}
	return err
}

// GetInstance gets a instance on Civo.
func (c *CivoClient) GetInstance(id string) (*civogo.Instance, error) {
	instance, err := c.CivoGoClient.GetInstance(id)
	if err != nil {
		if strings.Contains(err.Error(), "DatabaseInstanceNotFoundError") {
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

// GetK3sCluster gets a K3s cluster on Civo.
func (c *CivoClient) GetK3sCluster(clusterName string) (*civogo.KubernetesCluster, error) {

	kubernetesCluster, err := c.CivoGoClient.FindKubernetesCluster(clusterName)
	if err != nil {
		if strings.Contains(err.Error(), "ZeroMatchesError") {
			return nil, nil
		}
		return nil, err
	}
	return kubernetesCluster, nil
}

// CreateNewK3sCluster creates a new K3s cluster on Civo.
func (c *CivoClient) CreateNewK3sCluster(clusterName string,
	pools []civogo.KubernetesClusterPoolConfig, applications []string, cni *string, version *string) error {

	// Find the default network ID
	network, err := c.CivoGoClient.GetDefaultNetwork()
	if err != nil {
		return err
	}

	if len(pools) < 1 {
		return errors.New("pool is required for CivoKubernetes cluster creation")
	}
	// Currently we will only define the initial pool entries to be created with the cluster
	// This is due to limitations in the API
	var cp string
	if cni != nil {
		cp = *cni
	} else {
		cp = "flannel"
	}

	ver := "1.22.2-k3s1"
	if version != nil {
		ver = *version
	}

	cfg := &civogo.KubernetesClusterConfig{
		Region:            c.CivoGoClient.Region,
		Name:              clusterName,
		Tags:              defaultTags,
		NetworkID:         network.ID,
		KubernetesVersion: ver,
		Pools:             pools,
		Applications:      strings.Join(applications, ","),
		CNIPlugin:         cp,
	}

	kubernetesCluster, err := c.CivoGoClient.NewKubernetesClusters(cfg)
	if err != nil {
		return err
	}

	log.Debugf("Created Kubernetes cluster %s with %d node pools", kubernetesCluster.Name, len(pools))

	return nil
}

// UpdateK3sCluster updates a K3s cluster on Civo.
func (c *CivoClient) UpdateK3sCluster(desiredCluster *providerCivoCluster.CivoKubernetes,
	remoteCivoCluster *civogo.KubernetesCluster, provider *v1alpha1provider.ProviderConfig) error {

	_, err := c.CivoGoClient.UpdateKubernetesCluster(desiredCluster.Spec.Name,
		&civogo.KubernetesClusterConfig{
			Pools: desiredCluster.Spec.Pools,
		})

	return err
}

// UpdateK3sClusterVersion updates a K3s cluster version on Civo.
func (c *CivoClient) UpdateK3sClusterVersion(desiredCluster *providerCivoCluster.CivoKubernetes,
	remoteCivoCluster *civogo.KubernetesCluster, provider *v1alpha1provider.ProviderConfig) error {

	_, err := c.CivoGoClient.UpdateKubernetesCluster(desiredCluster.Spec.Name,
		&civogo.KubernetesClusterConfig{
			KubernetesVersion: *desiredCluster.Spec.Version,
		})

	return err
}

// DeleteK3sCluster deletes a k3s cluster on Civo.
func (c *CivoClient) DeleteK3sCluster(name string) error {

	kubernetesCluster, err := c.GetK3sCluster(name)
	if err != nil {
		return err
	}
	resp, err := c.CivoGoClient.DeleteKubernetesCluster(kubernetesCluster.ID)
	if err != nil {
		log.Debugf("error [%s %s %s %s]", resp.Result, resp.ErrorDetails, resp.ErrorCode, resp.ErrorReason)
	}
	return err
}
