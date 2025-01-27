package rke2k3s

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetMachineSelectorConfig is a function that will set the machine selector configurations in the main.tf file.
func SetMachineSelectorConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	machineSelectorBlock := rkeConfigBlockBody.AppendNewBlock(defaults.MachineSelectorConfig, nil)
	machineSelectorBlockBody := machineSelectorBlock.Body()

	registryValue := hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\n" + systemDefaultRegistry + ": " + terraformConfig.PrivateRegistries.SystemDefaultRegistry + "\nEOF"},
	})

	machineSelectorBlockBody.SetAttributeRaw(defaults.Config, registryValue)

	return nil
}

// SetPrivateRegistryConfig is a function that will set the private registry configurations in the main.tf file.
func SetPrivateRegistryConfig(registryBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	configBlock := registryBlockBody.AppendNewBlock(defaults.Configs, nil)
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(hostname, cty.StringVal(terraformConfig.PrivateRegistries.URL))

	if terraformConfig.PrivateRegistries.Username != "" {
		configBlockBody.SetAttributeValue(authConfigSecretName, cty.StringVal(terraformConfig.PrivateRegistries.AuthConfigSecretName))
	}

	configBlockBody.SetAttributeValue(tlsSecretName, cty.StringVal(terraformConfig.PrivateRegistries.TLSSecretName))
	configBlockBody.SetAttributeValue(caBundleName, cty.StringVal(terraformConfig.PrivateRegistries.CABundle))
	configBlockBody.SetAttributeValue(insecure, cty.BoolVal(terraformConfig.PrivateRegistries.Insecure))

	mirrorsBlock := registryBlockBody.AppendNewBlock(defaults.Mirrors, nil)
	mirrorsBlockBody := mirrorsBlock.Body()

	mirrorsBlockBody.SetAttributeValue(hostname, cty.StringVal(terraformConfig.PrivateRegistries.URL))

	return nil
}
