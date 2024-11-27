package rke1

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/instances"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/providers"
	"github.com/sirupsen/logrus"
)

// SetCustomRKE1 is a function that will set the custom RKE1 cluster configurations in the main.tf file.
func SetCustomRKE1(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, clusterConfig *config.TerratestConfig, configMap []map[string]any, clusterName string,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	if terraformConfig.MultiCluster {
		instances.SetAwsInstances(rootBody, terraformConfig, clusterConfig, clusterName)

		setRancher2Cluster(rootBody, terraformConfig, clusterName)

		nullresource.SetNullResource(rootBody, terraformConfig, clusterName)
	} else {
		providers.SetCustomProviders(rancherConfig, terraformConfig)

		instances.SetAwsInstances(rootBody, terraformConfig, clusterConfig, clusterName)

		setRancher2Cluster(rootBody, terraformConfig, clusterName)

		nullresource.SetNullResource(rootBody, terraformConfig, clusterName)

		locals.SetLocals(rootBody, terraformConfig, configMap, clusterName, newFile, file, nil)
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom RKE1 configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
