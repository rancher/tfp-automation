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

// CreateGoogleBackendService will set up the Google Cloud backend service.
func CreateGoogleBackendService(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	backendServiceBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeRegionBackendService, googleDefaults.GoogleComputeRegionBackendService + "_" + strconv.FormatInt(port, 10)})
	backendServiceBlockBody := backendServiceBlock.Body()

	backendServiceBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-backend-"+strconv.FormatInt(port, 10)))
	backendServiceBlockBody.SetAttributeValue(protocol, cty.StringVal("TCP"))
	backendServiceBlockBody.SetAttributeValue(timeoutSecond, cty.NumberIntVal(10))

	expression := "[" + googleDefaults.GoogleComputeRegionHealthCheck + "." + googleDefaults.GoogleComputeRegionHealthCheck + "_" + strconv.FormatInt(port, 10) + ".self_link]"
	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	backendServiceBlockBody.SetAttributeRaw(healthChecks, value)
	backendServiceBlockBody.SetAttributeValue(loadBalancerScheme, cty.StringVal("EXTERNAL"))
	backendServiceBlockBody.SetAttributeValue(googleDefaults.GoogleRegion, cty.StringVal(terraformConfig.GoogleConfig.Region))

	backendBlock := backendServiceBlockBody.AppendNewBlock(backend, nil)
	backendBlockBody := backendBlock.Body()

	expression = googleDefaults.GoogleComputeInstanceGroup + `.` + googleDefaults.GoogleComputeInstanceGroup + "_" + strconv.FormatInt(port, 10) + `.self_link`
	value = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	backendBlockBody.SetAttributeRaw(group, value)
	backendBlockBody.SetAttributeValue(balancingMode, cty.StringVal("CONNECTION"))
}
