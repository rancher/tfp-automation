package rke1

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
	"github.com/sirupsen/logrus"
)

// SetCustomRKE1 is a function that will set the custom RKE1 cluster configurations in the main.tf file.
func SetCustomRKE1(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any, clusterName string,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, clusterName)

	SetRancher2Cluster(rootBody, terraformConfig, terratestConfig, clusterName)

	nullresource.SetNullResource(rootBody, terraformConfig, clusterName)

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom RKE1 configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
