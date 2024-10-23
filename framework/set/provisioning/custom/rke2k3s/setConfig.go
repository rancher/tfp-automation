package rke2k3s

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/instances"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/sirupsen/logrus"
)

// SetCustomRKE2K3s is a function that will set the custom RKE2/K3s cluster configurations in the main.tf file.
func SetCustomRKE2K3s(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, clusterConfig *config.TerratestConfig, clusterName string,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) error {
	instances.SetAwsInstances(rootBody, terraformConfig, clusterConfig)

	setRancher2ClusterV2(rootBody, terraformConfig, clusterConfig, clusterName)

	nullresource.SetNullResource(rootBody, terraformConfig, clusterName)

	locals.SetLocals(rootBody, terraformConfig, clusterName)

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom RKE2/K3s configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}