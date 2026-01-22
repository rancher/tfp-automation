package google

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/zclconf/go-cty/cty"
)

// CreateGoogleHealthCheck will set up the Google Cloud health check.
func CreateGoogleHealthCheck(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	healthCheckBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeRegionHealthCheck, googleDefaults.GoogleComputeRegionHealthCheck + "_" + strconv.FormatInt(port, 10)})
	healthCheckBlockBody := healthCheckBlock.Body()

	healthCheckBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-health-check-"+strconv.FormatInt(port, 10)))

	tcpHealthCheckBlock := healthCheckBlockBody.AppendNewBlock(tcpHealthCheck, nil)
	tcpHealthCheckBlockBody := tcpHealthCheckBlock.Body()

	tcpHealthCheckBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
}
