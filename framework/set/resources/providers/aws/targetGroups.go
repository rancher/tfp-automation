package aws

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
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

// CreateTargetGroups is a function that will set the target group configurations in the main.tf file.
func CreateTargetGroups(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.LoadBalancerTargetGroup, aws.TargetGroupPrefix + strconv.FormatInt(port, 10)})
	targetGroupBlockBody := targetGroupBlock.Body()

	targetGroupBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
	targetGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(TCP))
	targetGroupBlockBody.SetAttributeValue(aws.VpcId, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	targetGroupBlockBody.SetAttributeValue(aws.TargetType, cty.StringVal(terraformConfig.AWSConfig.TargetType))
	targetGroupBlockBody.SetAttributeValue(aws.IPAddressType, cty.StringVal(terraformConfig.AWSConfig.IPAddressType))
	targetGroupBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.ResourcePrefix+"-tg-"+strconv.FormatInt(port, 10)))

	healthCheckGroupBlock := targetGroupBlockBody.AppendNewBlock(aws.HealthCheck, nil)
	healthCheckGroupBlockBody := healthCheckGroupBlock.Body()

	if terraformConfig.AWSConfig.IPAddressType == aws.IPv6 {
		healthCheckGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(TCP))
	} else {
		healthCheckGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(HTTP))
		healthCheckGroupBlockBody.SetAttributeValue(path, cty.StringVal(ping))
		healthCheckGroupBlockBody.SetAttributeValue(matcher, cty.StringVal("200-399"))
	}

	healthCheckGroupBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
	healthCheckGroupBlockBody.SetAttributeValue(interval, cty.NumberIntVal(10))
	healthCheckGroupBlockBody.SetAttributeValue(timeout, cty.NumberIntVal(6))
	healthCheckGroupBlockBody.SetAttributeValue(healthyThreshold, cty.NumberIntVal(3))
	healthCheckGroupBlockBody.SetAttributeValue(unhealthyThreshold, cty.NumberIntVal(3))
}

// CreateInternalTargetGroups is a function that will set the internal target group configurations in the main.tf file.
func CreateInternalTargetGroups(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.LoadBalancerTargetGroup, aws.TargetGroupInternalPrefix + strconv.FormatInt(port, 10)})
	targetGroupBlockBody := targetGroupBlock.Body()

	targetGroupBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
	targetGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(TCP))
	targetGroupBlockBody.SetAttributeValue(aws.VpcId, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	targetGroupBlockBody.SetAttributeValue(aws.TargetType, cty.StringVal(terraformConfig.AWSConfig.TargetType))
	targetGroupBlockBody.SetAttributeValue(aws.IPAddressType, cty.StringVal(terraformConfig.AWSConfig.IPAddressType))
	targetGroupBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.ResourcePrefix+"-internal-tg-"+strconv.FormatInt(port, 10)))

	healthCheckGroupBlock := targetGroupBlockBody.AppendNewBlock(aws.HealthCheck, nil)
	healthCheckGroupBlockBody := healthCheckGroupBlock.Body()

	healthCheckGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(HTTP))
	healthCheckGroupBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
	healthCheckGroupBlockBody.SetAttributeValue(path, cty.StringVal(ping))
	healthCheckGroupBlockBody.SetAttributeValue(interval, cty.NumberIntVal(10))
	healthCheckGroupBlockBody.SetAttributeValue(timeout, cty.NumberIntVal(6))
	healthCheckGroupBlockBody.SetAttributeValue(healthyThreshold, cty.NumberIntVal(3))
	healthCheckGroupBlockBody.SetAttributeValue(unhealthyThreshold, cty.NumberIntVal(3))
	healthCheckGroupBlockBody.SetAttributeValue(matcher, cty.StringVal("200-399"))
}
