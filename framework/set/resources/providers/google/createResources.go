package google

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

// CreateGoogleCloudResources is a helper function that will create the Google Cloud resources needed for the RKE2 cluster.
func CreateGoogleCloudResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateGoogleCloudTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateGoogleCloudProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateGoogleCloudFirewalls(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateGoogleCloudLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateGoogleCloudInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	if terraformConfig.Standalone.CertManagerVersion != "" {
		ports := []int64{80, 443, 6443, 9345}
		for _, port := range ports {
			CreateGoogleCloudInsanceGroups(rootBody, terraformConfig, port)
			rootBody.AppendNewline()

			CreateGoogleBackendService(rootBody, terraformConfig, port)
			rootBody.AppendNewline()

			CreateGoogleHealthCheck(rootBody, terraformConfig, port)
			rootBody.AppendNewline()

			CreateGoogleForwardingRule(rootBody, terraformConfig, port)
			rootBody.AppendNewline()
		}
	}

	CreateGoogleCloudLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}
