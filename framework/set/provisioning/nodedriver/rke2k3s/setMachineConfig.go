package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// setMachineConfig is a function that will set the machine configurations in the main.tf file.
func setMachineConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, psact string) (*hclwrite.Body, error) {
	machineConfigBlock := rootBody.AppendNewBlock(general.Resource, []string{machineConfigV2, terraformConfig.ResourcePrefix})
	machineConfigBlockBody := machineConfigBlock.Body()

	if psact == clusters.RancherBaseline {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + rancher2.PodSecurityAdmission + "." +
				terraformConfig.ResourcePrefix + "]")},
		}

		machineConfigBlockBody.SetAttributeRaw(general.DependsOn, dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue(general.GenerateName, cty.StringVal(terraformConfig.ResourcePrefix))

	return machineConfigBlockBody, nil
}
