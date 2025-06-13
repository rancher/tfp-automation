package imported

import (
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	extClusters "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/tests/actions/clusters"
	"github.com/rancher/tests/actions/upgradeinput"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/sirupsen/logrus"
)

// SetUpgradeImportedCluster is a function that will upgraded the imported cluster.
func SetUpgradeImportedCluster(client *rancher.Client, terraformConfig *config.TerraformConfig) error {
	clusterObj, err := extClusters.GetClusterIDByName(client, terraformConfig.ResourcePrefix)
	if err != nil {
		return err
	}

	clusterResp, err := client.Management.Cluster.ByID(clusterObj)
	if err != nil {
		return err
	}

	cluster, err := upgradeinput.LoadUpgradeKubernetesConfig(client)
	if err != nil {
		return err
	}

	clusterConfig := clusters.ConvertConfigToClusterConfig(&cluster[0].ProvisioningInput)
	clusterConfig.KubernetesVersion = cluster[0].VersionToUpgrade

	var updatedCluster *management.Cluster
	if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		clusterConfig.KubernetesVersion += "+k3s1"

		updatedCluster = &management.Cluster{
			K3sConfig: &management.K3sConfig{
				Version: clusterConfig.KubernetesVersion,
			},
			Name: clusterResp.Name,
		}
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		clusterConfig.KubernetesVersion += "+rke2r1"

		updatedCluster = &management.Cluster{
			Rke2Config: &management.Rke2Config{
				Version: clusterConfig.KubernetesVersion,
			},
			Name: clusterResp.Name,
		}
	}

	updatedClusterResp, err := extClusters.UpdateRKE1Cluster(client, clusterResp, updatedCluster)
	if err != nil {
		return err
	}

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		logrus.Infof("Cluster has been upgraded to: %s", updatedClusterResp.K3sConfig.Version)
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		logrus.Infof("Cluster has been upgraded to: %s", updatedClusterResp.Rke2Config.Version)
	}

	return nil
}
