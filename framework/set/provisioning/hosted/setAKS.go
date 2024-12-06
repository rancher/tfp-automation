package hosted

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/azure"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetAKS is a function that will set the AKS configurations in the main.tf file.
func SetAKS(terraformConfig *config.TerraformConfig, clusterName, k8sVersion string, nodePools []config.Nodepool, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	cloudCredBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.CloudCredential, defaults.CloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.CloudCredentialName))

	azCredConfigBlock := cloudCredBlockBody.AppendNewBlock(azure.AzureCredentialConfig, nil)
	azCredConfigBlockBody := azCredConfigBlock.Body()

	azCredConfigBlockBody.SetAttributeValue(azure.ClientID, cty.StringVal(terraformConfig.AzureCredentials.ClientID))
	azCredConfigBlockBody.SetAttributeValue(azure.ClientSecret, cty.StringVal(terraformConfig.AzureCredentials.ClientSecret))
	azCredConfigBlockBody.SetAttributeValue(azure.SubscriptionID, cty.StringVal(terraformConfig.AzureCredentials.SubscriptionID))
	azCredConfigBlockBody.SetAttributeValue(azure.TenantID, cty.StringVal(terraformConfig.AzureCredentials.TenantID))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, defaults.Cluster})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))

	aksConfigBlock := clusterBlockBody.AppendNewBlock(azure.AKSConfig, nil)
	aksConfigBlockBody := aksConfigBlock.Body()

	cloudCredID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.CloudCredential + "." + defaults.CloudCredential + ".id")},
	}

	aksConfigBlockBody.SetAttributeRaw(defaults.CloudCredentialID, cloudCredID)
	aksConfigBlockBody.SetAttributeValue(azure.OutboundType, cty.StringVal(terraformConfig.AzureConfig.OutboundType))
	aksConfigBlockBody.SetAttributeValue(azure.ResourceGroup, cty.StringVal(terraformConfig.AzureConfig.ResourceGroup))
	aksConfigBlockBody.SetAttributeValue(azure.ResourceLocation, cty.StringVal(terraformConfig.AzureConfig.ResourceLocation))
	aksConfigBlockBody.SetAttributeValue(azure.DNSPrefix, cty.StringVal(terraformConfig.HostnamePrefix))
	aksConfigBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(k8sVersion))
	aksConfigBlockBody.SetAttributeValue(azure.NetworkPlugin, cty.StringVal(terraformConfig.NetworkPlugin))
	aksConfigBlockBody.SetAttributeValue(azure.VirtualNetwork, cty.StringVal(terraformConfig.AzureConfig.Vnet))
	aksConfigBlockBody.SetAttributeValue(azure.Subnet, cty.StringVal(terraformConfig.AzureConfig.Subnet))
	aksConfigBlockBody.SetAttributeValue(azure.NetworkDNSServiceIP, cty.StringVal(terraformConfig.AzureConfig.NetworkDNSServiceIP))
	aksConfigBlockBody.SetAttributeValue(azure.NetworkDockerBridgeCIDR, cty.StringVal(terraformConfig.AzureConfig.NetworkDockerBridgeCIDR))
	aksConfigBlockBody.SetAttributeValue(azure.NetworkServiceCIDR, cty.StringVal(terraformConfig.AzureConfig.NetworkServiceCIDR))

	availabilityZones := format.ListOfStrings(terraformConfig.AzureConfig.AvailabilityZones)

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := resources.SetResourceNodepoolValidation(terraformConfig, pool, poolNum)
		if err != nil {
			return nil, err
		}

		nodePoolsBlock := aksConfigBlockBody.AppendNewBlock(azure.NodePools, nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeRaw(azure.AvailabilityZones, availabilityZones)
		nodePoolsBlockBody.SetAttributeValue(azure.NodePoolMode, cty.StringVal(terraformConfig.AzureConfig.Mode))
		nodePoolsBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.AzureConfig.Name))
		nodePoolsBlockBody.SetAttributeValue(azure.Count, cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue(azure.OrchestratorVersion, cty.StringVal(k8sVersion))
		nodePoolsBlockBody.SetAttributeValue(azure.OSDiskSizeGB, cty.NumberIntVal(terraformConfig.AzureConfig.OSDiskSizeGB))
		nodePoolsBlockBody.SetAttributeValue(azure.VMSize, cty.StringVal(terraformConfig.AzureConfig.VMSize))

		taints := format.ListOfStrings(terraformConfig.AzureConfig.Taints)
		nodePoolsBlockBody.SetAttributeRaw(azure.Taints, taints)
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write AKS configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
