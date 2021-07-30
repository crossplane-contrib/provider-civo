package civocli

import (
	"strings"

	"github.com/civo/civogo"
	"github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
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
	civoGoClient *civogo.Client
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
		civoGoClient: client,
	}, nil
}

// UpdateInstance updates a civo instance
func (c *CivoClient) UpdateInstance(id string, instance *v1alpha1.CivoInstance) error {
	civoInstance, err := c.civoGoClient.GetInstance(id)
	if err != nil {
		return err
	}
	if civoInstance.Hostname != instance.Spec.InstanceConfig.Hostname {
		civoInstance.Hostname = instance.Spec.InstanceConfig.Hostname
	}
	if civoInstance.Notes != instance.Spec.InstanceConfig.Notes {
		civoInstance.Notes = instance.Spec.InstanceConfig.Notes
	}
	_, err = c.civoGoClient.UpdateInstance(civoInstance)
	if err != nil {
		return err
	}
	return nil
}

// CreateNewInstance creates a new instance on Civo.
func (c *CivoClient) CreateNewInstance(instance *v1alpha1.CivoInstance, sshPubKey, diskImageName string) (*civogo.Instance, error) {
	config, err := c.civoGoClient.NewInstanceConfig()
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
		if sshKey, err := c.civoGoClient.FindSSHKey(config.Hostname); err == nil {
			config.SSHKeyID = sshKey.ID
		} else {
			newSSHKey, err := c.civoGoClient.NewSSHKey(config.Hostname, sshPubKey)
			if err != nil {
				return nil, err
			}
			config.SSHKeyID = newSSHKey.ID
		}
	}

	template, err := c.civoGoClient.FindDiskImage(diskImageName)
	if err != nil {
		return nil, err
	}
	config.TemplateID = template.ID
	result, err := c.civoGoClient.CreateInstance(config)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteInstance deletes a instance on Civo.
func (c *CivoClient) DeleteInstance(id string) error {
	instance, err := c.civoGoClient.GetInstance(id)
	if err != nil {
		return err
	}
	resp, err := c.civoGoClient.DeleteInstance(instance.ID)
	if err != nil {
		log.Debugf("error [%s %s %s %s]", resp.Result, resp.ErrorDetails, resp.ErrorCode, resp.ErrorReason)
	}
	if sshKey, err := c.civoGoClient.FindSSHKey(instance.Hostname); err == nil {
		_, err := c.civoGoClient.DeleteSSHKey(sshKey.ID)
		if err != nil {
			return err
		}
	}
	return err
}

// GetInstance gets a instance on Civo.
func (c *CivoClient) GetInstance(id string) (*civogo.Instance, error) {
	instance, err := c.civoGoClient.GetInstance(id)
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
