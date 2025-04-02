package linode

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

// CreateLinodeResources is a helper function that will create the Linode resources needed for the RKE2 cluster.
func CreateLinodeResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateLinodeTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateLinodeProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	if terraformConfig.Standalone.RancherHostname != "" {
		CreateNodeBalancer(rootBody, terraformConfig)
		rootBody.AppendNewline()

		ports := []int64{80, 443, 6443, 9345}
		for _, port := range ports {
			CreateNodeBalancerConfig(rootBody, terraformConfig, port)
			rootBody.AppendNewline()

			CreateNodeBalancerNode(rootBody, terraformConfig, port)
			rootBody.AppendNewline()
		}
	}

	for _, instance := range instances {
		CreateLinodeInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	CreateLinodeLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}
