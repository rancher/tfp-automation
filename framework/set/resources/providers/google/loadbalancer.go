package google

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/zclconf/go-cty/cty"
)

// CreateGoogleCloudLoadBalancer will set up the Google Cloud load balancer.
func CreateGoogleCloudLoadBalancer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	loadBalancerBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeAddress, googleDefaults.GoogleComputeAddress})
	loadBalancerBlockBody := loadBalancerBlock.Body()

	loadBalancerBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	loadBalancerBlockBody.SetAttributeValue(googleDefaults.GoogleRegion, cty.StringVal(terraformConfig.GoogleConfig.Region))

}
