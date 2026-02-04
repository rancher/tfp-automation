package azure

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureLoadBalancerRules is a function that will set the load balancer rules configurations in the main.tf file.
func CreateAzureLoadBalancerRules(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	probeBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureLoadBalancerProbe, azure.AzureLoadBalancerProbe + "-" + strconv.FormatInt(port, 10)})
	probeBlockBody := probeBlock.Body()

	probeBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-probe"+"-"+strconv.FormatInt(port, 10)))

	expression := azure.AzureLoadBalancer + "." + azure.AzureLoadBalancer + ".id"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	probeBlockBody.SetAttributeRaw(loadBalancerID, values)
	probeBlockBody.SetAttributeValue(protocol, cty.StringVal("Tcp"))
	probeBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
	probeBlockBody.SetAttributeValue(internalInSeconds, cty.NumberIntVal(5))
	probeBlockBody.SetAttributeValue(numberOfProbes, cty.NumberIntVal(2))

	rootBody.AppendNewline()

	ruleBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureLoadBalancerRule, azure.AzureLoadBalancerRule + "-" + strconv.FormatInt(port, 10)})
	ruleBlockBody := ruleBlock.Body()

	ruleBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-rule"+"-"+strconv.FormatInt(port, 10)))
	ruleBlockBody.SetAttributeRaw(loadBalancerID, values)
	ruleBlockBody.SetAttributeValue(protocol, cty.StringVal("Tcp"))
	ruleBlockBody.SetAttributeValue(frontendPort, cty.NumberIntVal(port))
	ruleBlockBody.SetAttributeValue(backendPort, cty.NumberIntVal(port))
	ruleBlockBody.SetAttributeValue(frontendIPConfig, cty.StringVal(frontend))

	expression = "[" + azure.AzureLoadBalancerBackendAddressPool + "." + azure.AzureLoadBalancerBackendAddressPool + ".id]"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	ruleBlockBody.SetAttributeRaw(backendAddressPoolIDs, values)

	expression = azure.AzureLoadBalancerProbe + "." + azure.AzureLoadBalancerProbe + "-" + strconv.FormatInt(port, 10) + ".id"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	ruleBlockBody.SetAttributeRaw(probeID, values)
}
