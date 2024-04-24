package provisioning

import (
	"os"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/sirupsen/logrus"
)

// SetConfigTF is a function that will set the main.tf file based on the module type.
func SetConfigTF(clusterConfig *config.TerratestConfig, clusterName, poolName string) error {
	rancherConfig := new(rancher.Config)
	framework.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig(configs.Terraform, terraformConfig)

	module := terraformConfig.Module

	var file *os.File
	keyPath := SetKeyPath()

	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return err
	}

	defer file.Close()

	switch {
	case module == clustertypes.AKS:
		err = SetAKS(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, file)
		return err
	case module == clustertypes.EKS:
		err = SetEKS(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, file)
		return err
	case module == clustertypes.GKE:
		err = SetGKE(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, file)
		return err
	case strings.Contains(module, clustertypes.RKE1):
		err = SetRKE1(clusterName, poolName, clusterConfig.KubernetesVersion, clusterConfig.PSACT, clusterConfig.Nodepools, clusterConfig.SnapshotInput, file)
		return err
	case strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S):
		err = SetRKE2K3s(clusterName, poolName, clusterConfig.KubernetesVersion, clusterConfig.PSACT, clusterConfig.Nodepools, clusterConfig.SnapshotInput, file)
		return err
	default:
		logrus.Errorf("Unsupported module: %v", module)
	}

	return nil
}
