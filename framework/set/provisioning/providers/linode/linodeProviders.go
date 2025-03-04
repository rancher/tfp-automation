package linode

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetLinodeRKE1Provider is a helper function that will set the Linode RKE1
// Terraform configurations in the main.tf file.
func SetLinodeRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeConfigBlock := nodeTemplateBlockBody.AppendNewBlock(linode.LinodeConfig, nil)
	linodeConfigBlockBody := linodeConfigBlock.Body()

	linodeConfigBlockBody.SetAttributeValue(linode.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))

	linodeConfigBlockBody.SetAttributeValue(linode.Image, cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	linodeConfigBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.LinodeConfig.Region))
	linodeConfigBlockBody.SetAttributeValue(linode.RootPass, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
}

// SetLinodeRKE2K3SProvider is a helper function that will set the Linode RKE2/K3S
// Terraform provider details in the main.tf file.
func SetLinodeRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	linodeCredBlock := cloudCredBlockBody.AppendNewBlock(linode.LinodeCredentialConfig, nil)
	linodeCredBlockBody := linodeCredBlock.Body()

	linodeCredBlockBody.SetAttributeValue(linode.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))
}
