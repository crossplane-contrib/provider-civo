package civocli

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"

	"github.com/civo/civogo"
	providerCivoCluster "github.com/crossplane-contrib/provider-civo/apis/civo/cluster/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
	v1alpha1objectstore "github.com/crossplane-contrib/provider-civo/apis/civo/objectstore/v1alpha1"
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

// GenerateObjectStoreObservation creates the CivoInstanceObservation from object store infos
func GenerateObjectStoreObservation(stats *civogo.ObjectStoreStats) (*v1alpha1objectstore.CivoObjectStoreObservation, error) {
	percentage := stats.SizeKBUtilised * 100 / stats.MaxSizeKB
	observation := v1alpha1objectstore.CivoObjectStoreObservation{
		UtilisedPercentage: strconv.FormatInt(percentage, 10),
		Conditions:         make([]metav1.Condition, 0),
	}
	return &observation, nil
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
func (c *CivoClient) CreateNewK3sCluster(clusterName string,
	pools []civogo.KubernetesClusterPoolConfig, applications []string, cni *string, version *string) error {

	// Find the default network ID
	network, err := c.civoGoClient.GetDefaultNetwork()
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
		Region:            c.civoGoClient.Region,
		Name:              clusterName,
		Tags:              defaultTags,
		NetworkID:         network.ID,
		KubernetesVersion: ver,
		Pools:             pools,
		Applications:      strings.Join(applications, ","),
		CNIPlugin:         cp,
	}

	kubernetesCluster, err := c.civoGoClient.NewKubernetesClusters(cfg)
	if err != nil {
		return err
	}

	log.Debugf("Created Kubernetes cluster %s with %d node pools", kubernetesCluster.Name, len(pools))

	return nil
}

// UpdateK3sCluster updates a K3s cluster on Civo.
func (c *CivoClient) UpdateK3sCluster(desiredCluster *providerCivoCluster.CivoKubernetes,
	remoteCivoCluster *civogo.KubernetesCluster, provider *v1alpha1provider.ProviderConfig) error {

	// Convert desiredCluster.Spec.Pools to the type expected by civogo package.
	convertedPools := ConvertKubernetesClusterPoolConfigs(desiredCluster.Spec.Pools)

	_, err := c.civoGoClient.UpdateKubernetesCluster(desiredCluster.Spec.Name,
		&civogo.KubernetesClusterConfig{
			Pools: convertedPools,
		})

	return err
}

// UpdateK3sClusterVersion updates a K3s cluster version on Civo.
func (c *CivoClient) UpdateK3sClusterVersion(desiredCluster *providerCivoCluster.CivoKubernetes,
	remoteCivoCluster *civogo.KubernetesCluster, provider *v1alpha1provider.ProviderConfig) error {

	_, err := c.civoGoClient.UpdateKubernetesCluster(desiredCluster.Spec.Name,
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
	resp, err := c.civoGoClient.DeleteKubernetesCluster(kubernetesCluster.ID)
	if err != nil {
		log.Debugf("error [%s %s %s %s]", resp.Result, resp.ErrorDetails, resp.ErrorCode, resp.ErrorReason)
	}
	return err
}

// GetObjectStore fetches a object store by its ID.
func (c *CivoClient) GetObjectStore(id string) (*civogo.ObjectStore, error) {
	objectStore, err := c.civoGoClient.GetObjectStore(id)
	if err != nil {
		if strings.Contains(err.Error(), "ZeroMatchesError") {
			return nil, nil
		}
		return nil, err
	}
	return objectStore, nil
}

// CreateObjectStore creates object store.
func (c *CivoClient) CreateObjectStore(name string, size int, accessKeyID string) (*civogo.ObjectStore, error) {
	objectStore, err := c.civoGoClient.NewObjectStore(&civogo.CreateObjectStoreRequest{
		Name:        name,
		MaxSizeGB:   int64(size),
		AccessKeyID: accessKeyID,
		Region:      c.civoGoClient.Region,
	})
	if err != nil {
		return nil, err
	}
	return objectStore, err
}

// UpdateObjectStore updates size of object store by its ID.
func (c *CivoClient) UpdateObjectStore(id string, size int) error {
	_, err := c.civoGoClient.UpdateObjectStore(id, &civogo.UpdateObjectStoreRequest{
		MaxSizeGB: int64(size),
		Region:    c.civoGoClient.Region,
	})
	return err
}

// DeleteObjectStore updates size of object store by its ID.
func (c *CivoClient) DeleteObjectStore(id string) error {
	_, err := c.civoGoClient.DeleteObjectStore(id)
	return err
}

// GetObjectStoreByName updates size of object store by its name.
func (c *CivoClient) GetObjectStoreByName(name string) (*civogo.ObjectStore, error) {
	allObjectStore, err := c.civoGoClient.ListObjectStores()
	if err != nil {
		return nil, err
	}

	for _, value := range allObjectStore.Items {
		if value.Name == name {
			return &value, nil
		}
	}
	return nil, errors.New("No such object store found")
}

// GetObjectStoreCredential fetches credentials of object store.
func (c *CivoClient) GetObjectStoreCredential(credentialID string) *civogo.ObjectStoreCredential {
	cred, err := c.civoGoClient.GetObjectStoreCredential(credentialID)
	if err != nil {
		return nil
	}
	return cred
}

// GetObjectStoreStats fetches stats of object store.
func (c *CivoClient) GetObjectStoreStats(id string) *civogo.ObjectStoreStats {
	stats, err := c.civoGoClient.GetObjectStoreStats(id)
	if err != nil {
		return nil
	}
	return stats
}
