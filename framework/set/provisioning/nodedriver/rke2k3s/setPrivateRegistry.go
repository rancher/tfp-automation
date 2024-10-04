package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setMachineSelectorConfig is a function that will set the machine selector configurations in the main.tf file.
func setMachineSelectorConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	machineSelectorBlock := rkeConfigBlockBody.AppendNewBlock(defaults.MachineSelectorConfig, nil)
	machineSelectorBlockBody := machineSelectorBlock.Body()

	configBlock := machineSelectorBlockBody.AppendNewBlock(defaults.Config+" =", nil)
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(systemDefaultRegistry, cty.StringVal(terraformConfig.PrivateRegistries.SystemDefaultRegistry))

	return nil
}

// setPrivateRegistryConfig is a function that will set the private registry configurations in the main.tf file.
func setPrivateRegistryConfig(registryBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	configBlock := registryBlockBody.AppendNewBlock(defaults.Configs, nil)
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(hostname, cty.StringVal(terraformConfig.PrivateRegistries.URL))
	configBlockBody.SetAttributeValue(authConfigSecretName, cty.StringVal(terraformConfig.PrivateRegistries.AuthConfigSecretName))
	configBlockBody.SetAttributeValue(tlsSecretName, cty.StringVal(terraformConfig.PrivateRegistries.TLSSecretName))
	configBlockBody.SetAttributeValue(caBundleName, cty.StringVal(terraformConfig.PrivateRegistries.CABundle))
	configBlockBody.SetAttributeValue(insecure, cty.BoolVal(terraformConfig.PrivateRegistries.Insecure))

	return nil
}
