package aws

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	internal = "internal"
	name     = "name"
	network  = "network"
)

// createLoadBalancer is a function that will set the load balancer configurations in the main.tf file.
func createLoadBalancer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	loadBalancerGroupBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LoadBalancer, defaults.LoadBalancer})
	loadBalancerGroupBodyBlockBody := loadBalancerGroupBlock.Body()

	loadBalancerGroupBodyBlockBody.SetAttributeValue(internal, cty.BoolVal(false))
	loadBalancerGroupBodyBlockBody.SetAttributeValue(defaults.LoadBalancerType, cty.StringVal(network))

	subnetList := format.ListOfStrings([]string{terraformConfig.AWSConfig.AWSSubnetID})
	loadBalancerGroupBodyBlockBody.SetAttributeRaw(defaults.Subnets, subnetList)
	loadBalancerGroupBodyBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.HostnamePrefix))
}
