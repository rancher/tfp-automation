package framework

import (
	"os"
	"strings"

	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/josh-diamond/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

const (
	aks        = "aks"
	eks        = "eks"
	gke        = "gke"
	ec2RKE1    = "ec2_rke1"
	ec2RKE2    = "ec2_rke2"
	ec2K3s     = "ec2_k3s"
	linodeRKE1 = "linode_rke1"
	linodeRKE2 = "linode_rke2"
	linodeK3s  = "linode_k3s"
	rke1       = "rke1"
	rke2       = "rke2"
	k3s        = "k3s"
)

// SetConfigTF is a function that will set the main.tf file based on the module type.
func SetConfigTF(clusterConfig *config.TerratestConfig, clusterName string) error {
	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig("terraform", terraformConfig)

	module := terraformConfig.Module

	var file *os.File
	keyPath := SetKeyPath()

	file, err := os.Create(keyPath + "/main.tf")
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return err
	}

	defer file.Close()

	switch {
	case module == aks:
		err = SetAKS(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, file)
		return err
	case module == eks:
		err = SetEKS(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, file)
		return err
	case module == gke:
		err = SetGKE(clusterName, clusterConfig.KubernetesVersion, clusterConfig.Nodepools, file)
		return err
	case strings.Contains(module, rke1):
		err = SetRKE1(clusterName, clusterConfig.KubernetesVersion, clusterConfig.PSACT, clusterConfig.Nodepools, file)
		return err
	case strings.Contains(module, rke2) || strings.Contains(module, k3s):
		err = SetRKE2K3s(clusterName, clusterConfig.KubernetesVersion, clusterConfig.PSACT, clusterConfig.Nodepools, file)
		return err
	default:
		logrus.Errorf("Unsupported module: %v", module)
	}

	return nil
}
