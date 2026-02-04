package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

const (
	access                   = "access"
	destinationAddressPrefix = "destination_address_prefix"
	destinationPortRange     = "destination_port_range"
	direction                = "direction"
	priority                 = "priority"
	protocol                 = "protocol"
	sourceAddressPrefix      = "source_address_prefix"
	sourcePortRange          = "source_port_range"
)

// CreateAzureNetworkSecurityGroup is a function that will set the network security group configurations in the main.tf file.
func CreateAzureNetworkSecurityGroup(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	nsgBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureNetworkSecurityGroup, azure.AzureNetworkSecurityGroup})
	nsgBlockBody := nsgBlock.Body()

	nsgBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-nsg"))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".location"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nsgBlockBody.SetAttributeRaw(location, values)

	expression = azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nsgBlockBody.SetAttributeRaw(resourceGroupName, values)

	securityRuleBlock := nsgBlockBody.AppendNewBlock(securityRule, nil)
	securityRuleBlockBody := securityRuleBlock.Body()

	securityRuleBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal("InboundRule"))
	securityRuleBlockBody.SetAttributeValue(priority, cty.NumberIntVal(100))
	securityRuleBlockBody.SetAttributeValue(direction, cty.StringVal("Inbound"))
	securityRuleBlockBody.SetAttributeValue(access, cty.StringVal("Allow"))
	securityRuleBlockBody.SetAttributeValue(protocol, cty.StringVal("Tcp"))
	securityRuleBlockBody.SetAttributeValue(sourcePortRange, cty.StringVal("*"))
	securityRuleBlockBody.SetAttributeValue(destinationPortRange, cty.StringVal("*"))
	securityRuleBlockBody.SetAttributeValue(sourceAddressPrefix, cty.StringVal("*"))
	securityRuleBlockBody.SetAttributeValue(destinationAddressPrefix, cty.StringVal("*"))
}
