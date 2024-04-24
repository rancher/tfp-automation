package provisioning

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	blocks "github.com/rancher/tfp-automation/defaults/resourceblocks"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/azure"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetAKS is a function that will set the AKS configurations in the main.tf file.
func SetAKS(clusterName, k8sVersion string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	framework.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig(configs.Terraform, terraformConfig)

	newFile, rootBody := SetProvidersAndUsersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock(blocks.Resource, []string{blocks.CloudCredential, blocks.CloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(terraformConfig.CloudCredentialName))

	azCredConfigBlock := cloudCredBlockBody.AppendNewBlock(azure.AzureCredentialConfig, nil)
	azCredConfigBlockBody := azCredConfigBlock.Body()

	azCredConfigBlockBody.SetAttributeValue(azure.ClientID, cty.StringVal(terraformConfig.AzureConfig.ClientID))
	azCredConfigBlockBody.SetAttributeValue(azure.ClientSecret, cty.StringVal(terraformConfig.AzureConfig.ClientSecret))
	azCredConfigBlockBody.SetAttributeValue(azure.SubscriptionID, cty.StringVal(terraformConfig.AzureConfig.SubscriptionID))
	azCredConfigBlockBody.SetAttributeValue(azure.TenantID, cty.StringVal(terraformConfig.AzureConfig.TenantID))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(blocks.Resource, []string{blocks.Cluster, blocks.Cluster})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(clusterName))

	aksConfigBlock := clusterBlockBody.AppendNewBlock(azure.AKSConfig, nil)
	aksConfigBlockBody := aksConfigBlock.Body()

	cloudCredID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(blocks.CloudCredential + "." + blocks.CloudCredential + ".id")},
	}

	aksConfigBlockBody.SetAttributeRaw(blocks.CloudCredentialID, cloudCredID)
	aksConfigBlockBody.SetAttributeValue(azure.ResourceGroup, cty.StringVal(terraformConfig.AzureConfig.ResourceGroup))
	aksConfigBlockBody.SetAttributeValue(azure.ResourceLocation, cty.StringVal(terraformConfig.AzureConfig.ResourceLocation))
	aksConfigBlockBody.SetAttributeValue(azure.DNSPrefix, cty.StringVal(terraformConfig.HostnamePrefix))
	aksConfigBlockBody.SetAttributeValue(blocks.KubernetesVersion, cty.StringVal(k8sVersion))
	aksConfigBlockBody.SetAttributeValue(azure.NetworkPlugin, cty.StringVal(terraformConfig.NetworkPlugin))

	availabilityZones := format.ListOfStrings(terraformConfig.AzureConfig.AvailabilityZones)

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolsBlock := aksConfigBlockBody.AppendNewBlock(azure.NodePools, nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeRaw(azure.AvailabilityZones, availabilityZones)
		nodePoolsBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(blocks.Pool+poolNum))
		nodePoolsBlockBody.SetAttributeValue(azure.Count, cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue(azure.OrchestratorVersion, cty.StringVal(k8sVersion))
		nodePoolsBlockBody.SetAttributeValue(azure.OSDiskSizeGB, cty.NumberIntVal(terraformConfig.AzureConfig.OSDiskSizeGB))
		nodePoolsBlockBody.SetAttributeValue(azure.VMSize, cty.StringVal(terraformConfig.AzureConfig.VMSize))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write AKS configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
