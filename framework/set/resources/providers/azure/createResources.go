package azure

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

// CreateAzureResources is a helper function that will create the Azure resources needed for the RKE2 cluster.
func CreateAzureResources(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error) {
	CreateAzureTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	CreateAzureProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateAzureResourceGroup(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateAzureVirtualNetwork(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateAzureSubnet(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateAzurePublicIP(rootBody, terraformConfig, loadBalancerIP)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateAzurePublicIP(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	CreateAzureLoadBalancer(rootBody, terraformConfig)
	rootBody.AppendNewline()

	CreateAzureLoadBalancerBackend(rootBody, terraformConfig)
	rootBody.AppendNewline()

	ports := []int64{80, 443, 6443, 9345}
	for _, port := range ports {
		CreateAzureLoadBalancerRules(rootBody, terraformConfig, port)
		rootBody.AppendNewline()
	}

	CreateAzureNetworkSecurityGroup(rootBody, terraformConfig)
	rootBody.AppendNewline()

	for _, instance := range instances {
		CreateAzureNetworkInterface(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()

		CreateAzureNetworkInterfaceSecurityGroupAssociation(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()

		CreateAzureNetworkInterfaceBackend(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()

		CreateAzureInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	CreateAzureLocalBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, err
}
