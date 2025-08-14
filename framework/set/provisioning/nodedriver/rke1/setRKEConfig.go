package rke1

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setRKEConfig is a function that will set the RKE configurations in the main.tf file.
func setRKEConfig(clusterBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, kubernetesVersion string) (*hclwrite.Body, error) {
	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(kubernetesVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(defaults.Plugin, cty.StringVal(terraformConfig.CNI))

	return rkeConfigBlockBody, nil
}
