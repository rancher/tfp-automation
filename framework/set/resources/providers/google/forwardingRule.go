package google

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/zclconf/go-cty/cty"
)

// CreateGoogleForwardingRule will set up the Google Cloud forwarding Rule resource.
func CreateGoogleForwardingRule(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	forwardingRuleBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeForwardingRule, googleDefaults.GoogleComputeForwardingRule + "_" + strconv.FormatInt(port, 10)})
	forwardingRuleBlockBody := forwardingRuleBlock.Body()

	forwardingRuleBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-forwarding-rule-"+strconv.FormatInt(port, 10)))
	forwardingRuleBlockBody.SetAttributeValue(googleDefaults.GoogleRegion, cty.StringVal(terraformConfig.GoogleConfig.Region))
	forwardingRuleBlockBody.SetAttributeValue(loadBalancerScheme, cty.StringVal("EXTERNAL"))

	expression := googleDefaults.GoogleComputeRegionBackendService + `.` + googleDefaults.GoogleComputeRegionBackendService + `_` + strconv.FormatInt(port, 10) + `.self_link`
	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	forwardingRuleBlockBody.SetAttributeRaw(backendService, value)
	forwardingRuleBlockBody.SetAttributeValue(ipProtocol, cty.StringVal("TCP"))
	forwardingRuleBlockBody.SetAttributeValue(general.Ports, cty.ListVal([]cty.Value{cty.StringVal(strconv.FormatInt(port, 10))}))

	expression = googleDefaults.GoogleComputeAddress + `.` + googleDefaults.GoogleComputeAddress + `.address`
	value = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	forwardingRuleBlockBody.SetAttributeRaw("ip_address", value)
}
