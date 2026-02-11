package azure

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzurePublicIP is a function that will set the public IP configurations in the main.tf file.
func CreateAzurePublicIP(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	publicIPBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzurePublicIP, azure.AzurePublicIP + "-" + hostnamePrefix})
	publicIPBlockBody := publicIPBlock.Body()

	publicIPBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-"+hostnamePrefix))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".location"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	publicIPBlockBody.SetAttributeRaw(location, values)

	expression = azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	publicIPBlockBody.SetAttributeRaw(resourceGroupName, values)
	publicIPBlockBody.SetAttributeValue(allocationMethod, cty.StringVal("Static"))
	publicIPBlockBody.SetAttributeValue(sku, cty.StringVal("Standard"))

	if strings.Contains(hostnamePrefix, "load") && terraformConfig.Provider != providers.AKS {
		publicIPBlockBody.SetAttributeValue(domainNameLabel, cty.StringVal(terraformConfig.ResourcePrefix))
	}
}
