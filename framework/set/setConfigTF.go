package set

import (
	"os"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	custom "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke1"
	customV2 "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/provisioning/hosted"
	nodedriver "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke1"
	nodedriverV2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/sirupsen/logrus"
)

// ConfigTF is a function that will set the main.tf file based on the module type.
func ConfigTF(client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, clusterConfig *config.TerratestConfig, clusterName, poolName string, rbacRole config.Role) error {

	module := terraformConfig.Module

	var file *os.File
	keyPath := resources.SetKeyPath()

	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return err
	}

	defer file.Close()

	newFile, rootBody := resources.SetProvidersAndUsersTF(rancherConfig, terraformConfig, false)

	rootBody.AppendNewline()

	switch {
	case module == clustertypes.AKS:
		err = hosted.SetAKS(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, newFile, rootBody, file)
		return err
	case module == clustertypes.EKS:
		err = hosted.SetEKS(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, newFile, rootBody, file)
		return err
	case module == clustertypes.GKE:
		err = hosted.SetGKE(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, newFile, rootBody, file)
		return err
	case strings.Contains(module, clustertypes.RKE1) && !strings.Contains(module, defaults.Custom):
		err = nodedriver.SetRKE1(clusterName, poolName, clusterConfig.KubernetesVersion, clusterConfig.PSACT, clusterConfig.Nodepools,
			clusterConfig.SnapshotInput, newFile, rootBody, file, rbacRole)
		return err
	case (strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S)) && !strings.Contains(module, defaults.Custom):
		err = nodedriverV2.SetRKE2K3s(client, clusterName, poolName, clusterConfig.KubernetesVersion, clusterConfig.PSACT, clusterConfig.Nodepools,
			clusterConfig.SnapshotInput, newFile, rootBody, file, rbacRole)
		return err
	case module == modules.CustomEC2RKE1:
		err = custom.SetCustomRKE1(rancherConfig, terraformConfig, clusterConfig, clusterName, newFile, rootBody, file)
		return err
	case module == modules.CustomEC2RKE2 || module == modules.CustomEC2K3s:
		err = customV2.SetCustomRKE2K3s(rancherConfig, terraformConfig, clusterConfig, clusterName, newFile, rootBody, file)
		return err
	default:
		logrus.Errorf("Unsupported module: %v", module)
	}

	return nil
}
