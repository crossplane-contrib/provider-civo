package civocli

import (
	"github.com/civo/civogo"
	"github.com/crossplane-contrib/provider-civo/apis/civo/cluster/v1alpha1"
)

// ConvertKubernetesClusterPoolConfigs converts a slice of KubernetesClusterPoolConfig from the
// provider-civo package to a slice of KubernetesClusterPoolConfig from the civogo package.
func ConvertKubernetesClusterPoolConfigs(pools []v1alpha1.KubernetesClusterPoolConfig) []civogo.KubernetesClusterPoolConfig {
	convertedPools := make([]civogo.KubernetesClusterPoolConfig, len(pools))
	for i, pool := range pools {
		convertedPools[i] = civogo.KubernetesClusterPoolConfig{
			Region:           pool.Region,
			ID:               pool.ID,
			Count:            pool.Count,
			Size:             pool.Size,
			Labels:           pool.Labels,
			Taints:           pool.Taints,
			PublicIPNodePool: pool.PublicIPNodePool,
		}
	}
	return convertedPools
}
