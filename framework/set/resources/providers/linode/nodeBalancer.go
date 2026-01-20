package linode

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	linodeDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
	"github.com/zclconf/go-cty/cty"
)

const (
	internal = "internal"
	name     = "name"
	network  = "network"
)

// CreateNodeBalancer is a function that will set the node balancer configurations in the main.tf file.
func CreateNodeBalancer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	nodeBalancerBlock := rootBody.AppendNewBlock(general.Resource, []string{linodeDefaults.LinodeNodeBalancer, linodeDefaults.LinodeNodeBalancer})
	nodeBalancerBlockBody := nodeBalancerBlock.Body()

	nodeBalancerBlockBody.SetAttributeValue(linode.Label, cty.StringVal(terraformConfig.ResourcePrefix))
	nodeBalancerBlockBody.SetAttributeValue(linode.Region, cty.StringVal(terraformConfig.LinodeConfig.Region))
	nodeBalancerBlockBody.SetAttributeValue(linode.ClientConnThrottle, cty.NumberIntVal(terraformConfig.LinodeConfig.ClientConnThrottle))

	tags := format.ListOfStrings(terraformConfig.LinodeConfig.Tags)
	nodeBalancerBlockBody.SetAttributeRaw(linode.Tags, tags)
}
