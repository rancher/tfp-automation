package multiclusters

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	configuration "github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	custom "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke1"
	customV2 "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/provisioning/hosted"
	nodedriver "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke1"
	nodedriverV2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/sirupsen/logrus"
)

// SetMultiCluster is a function that will set multiple cluster configurations in the main.tf file.
func SetMultiCluster(client *rancher.Client, rancherConfig *rancher.Config, configMap []map[string]any, clusterName string, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, rbacRole configuration.Role, poolName string) error {

	var err error
	customClusterNames := []string{}

	for i, clusterConfig := range configMap {

		terraformConfig := clusterConfig["terraform"].(configuration.TerraformConfig)
		terratestConfig := clusterConfig["terratest"].(configuration.TerratestConfig)

		kubernetesVersion := terratestConfig.KubernetesVersion
		nodePools := terratestConfig.Nodepools
		psact := terratestConfig.PSACT
		snapshotInput := terratestConfig.SnapshotInput

		module := terraformConfig.Module

		clusterName = namegen.AppendRandomString(configs.TFP)
		terraformConfig.ClusterName = clusterName
		poolName = namegen.AppendRandomString(configs.TFP)

		if terraformConfig.Module == modules.CustomEC2RKE2 || terraformConfig.Module == modules.CustomEC2K3s {
			customClusterNames = append(customClusterNames, clusterName)
		}

		switch {
		case module == clustertypes.AKS:
			file, err = hosted.SetAKS(&terraformConfig, clusterName, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return err
			}
		case module == clustertypes.EKS:
			file, err = hosted.SetEKS(&terraformConfig, clusterName, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return err
			}
		case module == clustertypes.GKE:
			file, err = hosted.SetGKE(&terraformConfig, clusterName, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return err
			}
		case strings.Contains(module, clustertypes.RKE1) && !strings.Contains(module, defaults.Custom):
			file, err = nodedriver.SetRKE1(&terraformConfig, clusterName, poolName, kubernetesVersion, psact, nodePools,
				snapshotInput, newFile, rootBody, file, rbacRole)
			if err != nil {
				return err
			}
		case (strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S)) && !strings.Contains(module, defaults.Custom):
			file, err = nodedriverV2.SetRKE2K3s(client, &terraformConfig, clusterName, poolName, kubernetesVersion, psact, nodePools,
				snapshotInput, newFile, rootBody, file, rbacRole)
			if err != nil {
				return err
			}
		case module == modules.CustomEC2RKE1:
			file, err = custom.SetCustomRKE1(rancherConfig, &terraformConfig, &terratestConfig, configMap, clusterName, newFile, rootBody, file)
			if err != nil {
				return err
			}
		case module == modules.CustomEC2RKE2 || module == modules.CustomEC2K3s:
			file, err = customV2.SetCustomRKE2K3s(rancherConfig, &terraformConfig, &terratestConfig, configMap, clusterName, newFile, rootBody, file)
			if err != nil {
				return err
			}
		default:
			logrus.Errorf("Unsupported module: %v", module)
		}

		if i == len(configMap)-1 {
			file, err = locals.SetLocals(rootBody, &terraformConfig, configMap, clusterName, newFile, file, customClusterNames)
		}
	}

	keyPath := resources.SetKeyPath()
	file, err = os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
