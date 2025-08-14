package rke1

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setRKE1PrivateRegistryConfig is a function that will set the private registry configurations in the main.tf file.
func setRKE1PrivateRegistryConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	registryBlock := rkeConfigBlockBody.AppendNewBlock(defaults.RKE1PrivateRegistries, nil)
	registryBlockBody := registryBlock.Body()

	registryBlockBody.SetAttributeValue(privateRegistryURL, cty.StringVal(terraformConfig.PrivateRegistries.URL))

	if terraformConfig.StandaloneRegistry.Authenticated {
		registryBlockBody.SetAttributeValue(privateRegistryUsername, cty.StringVal(terraformConfig.PrivateRegistries.Username))
		registryBlockBody.SetAttributeValue(privateRegistryPassword, cty.StringVal(terraformConfig.PrivateRegistries.Password))
	}

	return nil
}
