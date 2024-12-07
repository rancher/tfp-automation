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
func SetCustomRKE2K3s(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any, clusterName string,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	if terraformConfig.MultiCluster {
		instances.SetAwsInstances(rootBody, terraformConfig, terratestConfig, clusterName)

		setRancher2ClusterV2(rootBody, terraformConfig, terratestConfig, clusterName)

		nullresource.SetNullResource(rootBody, terraformConfig, clusterName)
	} else {
		instances.SetAwsInstances(rootBody, terraformConfig, terratestConfig, clusterName)

		setRancher2ClusterV2(rootBody, terraformConfig, terratestConfig, clusterName)

		nullresource.SetNullResource(rootBody, terraformConfig, clusterName)

		file, _ = locals.SetLocals(rootBody, terraformConfig, configMap, clusterName, newFile, file, nil)
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom RKE2/K3s configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
