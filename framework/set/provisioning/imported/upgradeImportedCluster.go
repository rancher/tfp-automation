package imported

import (
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	extClusters "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
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

	var clusterType string
	if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		clusterType = extClusters.K3SClusterType.String()
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		clusterType = extClusters.RKE2ClusterType.String()
	}

	version, err := kubernetesversions.Default(client, clusterType, nil)
	if err != nil {
		return err
	}

	var updatedCluster *management.Cluster
	if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		updatedCluster = &management.Cluster{
			K3sConfig: &management.K3sConfig{
				Version: version[0],
			},
			Name: clusterResp.Name,
		}
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		updatedCluster = &management.Cluster{
			Rke2Config: &management.Rke2Config{
				Version: version[0],
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
