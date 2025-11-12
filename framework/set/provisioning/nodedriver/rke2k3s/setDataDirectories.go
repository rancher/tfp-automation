package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetDataDirectories is a function that will set the data directories configurations in the main.tf file.
func SetDataDirectories(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body) error {

	dataDirectoriesBlock := rkeConfigBlockBody.AppendNewBlock(defaults.DataDirectories, nil)
	dataDirectoriesBlockBody := dataDirectoriesBlock.Body()

	dataDirectoriesBlockBody.SetAttributeValue(defaults.SystemAgent, cty.StringVal(terraformConfig.DataDirectories.SystemAgentPath))
	dataDirectoriesBlockBody.SetAttributeValue(defaults.Provisioning, cty.StringVal(terraformConfig.DataDirectories.ProvisioningPath))
	dataDirectoriesBlockBody.SetAttributeValue(defaults.K8sDistro, cty.StringVal(terraformConfig.DataDirectories.K8sDistroPath))

	return nil
}
