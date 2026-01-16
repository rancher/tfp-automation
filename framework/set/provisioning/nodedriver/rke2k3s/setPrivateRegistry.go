package rke2k3s

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// SetMachineSelectorConfig is a function that will set the machine selector configurations in the main.tf file.
func SetMachineSelectorConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	machineSelectorBlock := rkeConfigBlockBody.AppendNewBlock(clusters.MachineSelectorConfig, nil)
	machineSelectorBlockBody := machineSelectorBlock.Body()

	registryValue := hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\n" + systemDefaultRegistry + ": " + terraformConfig.PrivateRegistries.SystemDefaultRegistry + "\nEOF"},
	})

	machineSelectorBlockBody.SetAttributeRaw(clusters.Config, registryValue)

	return nil
}

// SetPrivateRegistryConfig is a function that will set the private registry configurations in the main.tf file.
func SetPrivateRegistryConfig(rkeConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	registryBlock := rkeConfigBlockBody.AppendNewBlock(clusters.PrivateRegistries, nil)
	registryBlockBody := registryBlock.Body()

	configBlock := registryBlockBody.AppendNewBlock(clusters.Configs, nil)
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(hostname, cty.StringVal(terraformConfig.PrivateRegistries.URL))

	if terraformConfig.PrivateRegistries.Username != "" {
		registrySecretName := terraformConfig.PrivateRegistries.AuthConfigSecretName + "-" + terraformConfig.ResourcePrefix
		configBlockBody.SetAttributeValue(authConfigSecretName, cty.StringVal(registrySecretName))
	}

	configBlockBody.SetAttributeValue(tlsSecretName, cty.StringVal(terraformConfig.PrivateRegistries.TLSSecretName))
	configBlockBody.SetAttributeValue(caBundleName, cty.StringVal(terraformConfig.PrivateRegistries.CABundle))
	configBlockBody.SetAttributeValue(insecure, cty.BoolVal(terraformConfig.PrivateRegistries.Insecure))

	mirrorsBlock := registryBlockBody.AppendNewBlock(clusters.Mirrors, nil)
	mirrorsBlockBody := mirrorsBlock.Body()

	mirrorsBlockBody.SetAttributeValue(hostname, cty.StringVal(terraformConfig.PrivateRegistries.MirrorHostname))
	mirrorsBlockBody.SetAttributeValue(endpoints, cty.ListVal([]cty.Value{cty.StringVal(terraformConfig.PrivateRegistries.MirrorEndpoint)}))

	return nil
}
