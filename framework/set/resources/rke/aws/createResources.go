package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
)

const (
	cloudProviderEnvVar = "CLOUD_PROVIDER_VERSION"
	rc                  = "-rc"
	requiredProviders   = "required_providers"
	rkeProviderEnvVar   = "RKE_PROVIDER_VERSION"
	rkeServerOne        = "rke_server1"
	rkeServerTwo        = "rke_server2"
	rkeServerThree      = "rke_server3"
	rkeLogFile          = "rke_debug.log"
)

// CreateAWSResources is a helper function that will create the AWS resources needed for the RKE1 cluster.
func CreateAWSResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig) (*os.File, error) {
	createTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	aws.CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	createRKEProviderBlock(rootBody)
	rootBody.AppendNewline()

	instances := []string{rkeServerOne, rkeServerTwo, rkeServerThree}

	for _, instance := range instances {
		aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}
