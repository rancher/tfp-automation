package aws

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	healthyThreshold   = "healthy_threshold"
	HTTP               = "HTTP"
	interval           = "interval"
	matcher            = "matcher"
	path               = "path"
	ping               = "/ping"
	protocol           = "protocol"
	TCP                = "TCP"
	timeout            = "timeout"
	trafficPort        = "traffic-port"
	unhealthyThreshold = "unhealthy_threshold"
)

// createTargetGroups is a function that will set the target group configurations in the main.tf file.
func createTargetGroups(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LoadBalancerTargetGroup, defaults.TargetGroupPrefix + strconv.FormatInt(port, 10)})
	targetGroupBlockBody := targetGroupBlock.Body()

	targetGroupBlockBody.SetAttributeValue(defaults.Port, cty.NumberIntVal(port))
	targetGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(TCP))
	targetGroupBlockBody.SetAttributeValue(defaults.VpcId, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	targetGroupBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.HostnamePrefix+"-tg-"+strconv.FormatInt(port, 10)))

	healthCheckGroupBlock := targetGroupBlockBody.AppendNewBlock(defaults.HealthCheck, nil)
	healthCheckGroupBlockBody := healthCheckGroupBlock.Body()

	healthCheckGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(HTTP))
	healthCheckGroupBlockBody.SetAttributeValue(defaults.Port, cty.StringVal(trafficPort))
	healthCheckGroupBlockBody.SetAttributeValue(path, cty.StringVal(ping))
	healthCheckGroupBlockBody.SetAttributeValue(interval, cty.NumberIntVal(10))
	healthCheckGroupBlockBody.SetAttributeValue(timeout, cty.NumberIntVal(6))
	healthCheckGroupBlockBody.SetAttributeValue(healthyThreshold, cty.NumberIntVal(3))
	healthCheckGroupBlockBody.SetAttributeValue(unhealthyThreshold, cty.NumberIntVal(3))
	healthCheckGroupBlockBody.SetAttributeValue(matcher, cty.StringVal("200-399"))
}
