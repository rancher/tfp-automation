package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setNetworkingConfig is a function that will set the networking configurations in the main.tf file.
func SetNetworkingConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	networkingConfigBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Networking, nil)
	networkingConfigBlockBody := networkingConfigBlock.Body()

	networkingConfigBlockBody.SetAttributeValue(defaults.StackPreference, cty.StringVal(terraformConfig.AWSConfig.Networking.StackPreference))

	return nil
}
