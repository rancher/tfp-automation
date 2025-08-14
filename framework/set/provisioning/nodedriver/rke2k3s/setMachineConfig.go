package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setMachineConfig is a function that will set the machine configurations in the main.tf file.
func setMachineConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, psact string) (*hclwrite.Body, hclwrite.Tokens, error) {
	machineConfigBlock := rootBody.AppendNewBlock(defaults.Resource, []string{machineConfigV2, terraformConfig.ResourcePrefix})
	machineConfigBlockBody := machineConfigBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Rancher2 + "." + defaults.StandardUser)},
	}

	machineConfigBlockBody.SetAttributeRaw(defaults.Provider, provider)

	if psact == defaults.RancherBaseline {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + defaults.PodSecurityAdmission + "." +
				terraformConfig.ResourcePrefix + "]")},
		}

		machineConfigBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue(defaults.GenerateName, cty.StringVal(terraformConfig.ResourcePrefix))

	return machineConfigBlockBody, provider, nil
}
