package aws

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/zclconf/go-cty/cty"
)

const (
	httpProtocolIPv6 = "http_protocol_ipv6"
	internal         = "internal"
	metadataOptions  = "metadata_options"
	name             = "name"
	network          = "network"
)

// CreateLoadBalancer is a function that will set the load balancer configurations in the main.tf file.
func CreateLoadBalancer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	loadBalancerBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.LoadBalancer, aws.LoadBalancer})
	loadBalancerBlockBody := loadBalancerBlock.Body()

	loadBalancerBlockBody.SetAttributeValue(internal, cty.BoolVal(false))
	loadBalancerBlockBody.SetAttributeValue(aws.LoadBalancerType, cty.StringVal(network))
	loadBalancerBlockBody.SetAttributeValue(aws.IPAddressType, cty.StringVal(terraformConfig.AWSConfig.LoadBalancerType))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	loadBalancerBlockBody.SetAttributeRaw(aws.SecurityGroups, securityGroups)

	subnetList := format.ListOfStrings([]string{terraformConfig.AWSConfig.AWSSubnetID})
	loadBalancerBlockBody.SetAttributeRaw(aws.Subnets, subnetList)
	loadBalancerBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.ResourcePrefix))
}

// CreateInternalLoadBalancer is a function that will set the internal load balancer configurations in the main.tf file.
func CreateInternalLoadBalancer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	loadBalancerBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.LoadBalancer, aws.InternalLoadBalancer})
	loadBalancerBlockBody := loadBalancerBlock.Body()

	loadBalancerBlockBody.SetAttributeValue(internal, cty.BoolVal(true))
	loadBalancerBlockBody.SetAttributeValue(aws.LoadBalancerType, cty.StringVal(network))
	loadBalancerBlockBody.SetAttributeValue(aws.IPAddressType, cty.StringVal(terraformConfig.AWSConfig.LoadBalancerType))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	loadBalancerBlockBody.SetAttributeRaw(aws.SecurityGroups, securityGroups)

	subnetList := format.ListOfStrings([]string{terraformConfig.AWSConfig.AWSSubnetID})
	loadBalancerBlockBody.SetAttributeRaw(aws.Subnets, subnetList)
	loadBalancerBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.ResourcePrefix+"-"+internal))
}
