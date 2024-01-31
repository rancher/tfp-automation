package linode

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// SetLinodeRKE1Provider is a helper function that will set the Linode RKE1 Terraform configurations in the main.tf file.
func SetLinodeRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeConfigBlock := nodeTemplateBlockBody.AppendNewBlock("linode_config", nil)
	linodeConfigBlockBody := linodeConfigBlock.Body()

	linodeConfigBlockBody.SetAttributeValue("image", cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	linodeConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.LinodeConfig.Region))
	linodeConfigBlockBody.SetAttributeValue("root_pass", cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	linodeConfigBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeConfig.LinodeToken))
}

// SetLinodeRKE2K3SProvider is a helper function that will set the Linode RKE2/K3S Terraform provider details in the main.tf file.
func SetLinodeRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	linodeCredBlock := cloudCredBlockBody.AppendNewBlock("linode_credential_config", nil)
	linodeCredBlockBody := linodeCredBlock.Body()

	linodeCredBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeConfig.LinodeToken))
}
