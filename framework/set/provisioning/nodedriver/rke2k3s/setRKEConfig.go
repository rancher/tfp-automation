package rke2k3s

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
)

// setRKEConfig is a function that will set the RKE configurations in the main.tf file.
func setRKEConfig(clusterBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) (*hclwrite.Body, error) {
	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	if terraformConfig.ChartValues != "" {
		chartValues := hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOF\n" + terraformConfig.ChartValues + "\nEOF"},
		})

		rkeConfigBlockBody.SetAttributeRaw(defaults.ChartValues, chartValues)
	}

	machineGlobalConfigValue := hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\ncni: " + terraformConfig.CNI + "\ndisable-kube-proxy: " + terraformConfig.DisableKubeProxy + "\nEOF"},
	})

	rkeConfigBlockBody.SetAttributeRaw(defaults.MachineGlobalConfig, machineGlobalConfigValue)

	return rkeConfigBlockBody, nil
}
