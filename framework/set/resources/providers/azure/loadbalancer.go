package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureLoadBalancer is a function that will set the load balancer configurations in the main.tf file.
func CreateAzureLoadBalancer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	lbBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureLoadBalancer, azure.AzureLoadBalancer})
	lbBlockBody := lbBlock.Body()

	lbBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-lb"))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".location"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	lbBlockBody.SetAttributeRaw("location", values)

	expression = azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	lbBlockBody.SetAttributeRaw(resourceGroupName, values)
	lbBlockBody.SetAttributeValue(sku, cty.StringVal("Standard"))

	frontendIPConfigBlock := lbBlockBody.AppendNewBlock("frontend_ip_configuration", nil)
	frontendIPConfigBlockBody := frontendIPConfigBlock.Body()

	frontendIPConfigBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(frontend))

	expression = azure.AzurePublicIP + "." + azure.AzurePublicIP + "-" + loadBalancerIP + ".id"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	frontendIPConfigBlockBody.SetAttributeRaw(publicIPAddressID, values)
}
