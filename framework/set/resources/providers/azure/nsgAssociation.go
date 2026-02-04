package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
)

// CreateAzureNetworkInterfaceSecurityGroupAssociation is a function that will set the network interface security associations
// configurations in the main.tf file.
func CreateAzureNetworkInterfaceSecurityGroupAssociation(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	nsgAssociationBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureNetworkSecurityGroupAssociation, azure.AzureNetworkSecurityGroupAssociation + "-" + hostnamePrefix})
	nsgAssociationBlockBody := nsgAssociationBlock.Body()

	expression := azure.AzureNetworkInterface + "." + azure.AzureNetworkInterface + "-" + hostnamePrefix + ".id"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nsgAssociationBlockBody.SetAttributeRaw(networkInterfaceID, values)

	expression = azure.AzureNetworkSecurityGroup + "." + azure.AzureNetworkSecurityGroup + ".id"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nsgAssociationBlockBody.SetAttributeRaw(networkSecurityGroupID, values)
}
