package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureNetworkInterfaceBackend is a function that will set the network interface backend configurations in the main.tf file.
func CreateAzureNetworkInterfaceBackend(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	networkInterfaceBackend := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureNetworkInterfaceBackendAddressPoolAssociation, azure.AzureNetworkInterfaceBackendAddressPoolAssociation + "-" + hostnamePrefix})
	networkInterfaceBackendBody := networkInterfaceBackend.Body()

	expression := azure.AzureNetworkInterface + "." + azure.AzureNetworkInterface + "-" + hostnamePrefix + ".id"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	networkInterfaceBackendBody.SetAttributeRaw(networkInterfaceID, values)
	networkInterfaceBackendBody.SetAttributeValue(ipConfigurationName, cty.StringVal(internal))

	expression = azure.AzureLoadBalancerBackendAddressPool + "." + azure.AzureLoadBalancerBackendAddressPool + ".id"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	networkInterfaceBackendBody.SetAttributeRaw(backendAddressPoolID, values)
}
