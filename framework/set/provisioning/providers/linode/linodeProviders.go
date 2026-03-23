package linode

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// SetLinodeRKE2K3SProvider is a helper function that will set the Linode RKE2/K3S
// Terraform provider details in the main.tf file.
func SetLinodeRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	linodeCredBlock := cloudCredBlockBody.AppendNewBlock(linode.LinodeCredentialConfig, nil)
	linodeCredBlockBody := linodeCredBlock.Body()

	linodeCredBlockBody.SetAttributeValue(linode.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))
}
