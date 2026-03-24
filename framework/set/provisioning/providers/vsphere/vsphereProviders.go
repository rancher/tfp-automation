package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/vsphere"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// SetVsphereRKE2K3SProvider is a helper function that will set the Vsphere RKE2/K3S Terraform provider details in the main.tf file.
func SetVsphereRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	vsphereCredBlock := cloudCredBlockBody.AppendNewBlock(vsphere.VsphereCredentialConfig, nil)
	vsphereCredBlockBody := vsphereCredBlock.Body()

	vsphereCredBlockBody.SetAttributeValue(vsphere.Password, cty.StringVal(terraformConfig.VsphereCredentials.Password))
	vsphereCredBlockBody.SetAttributeValue(vsphere.Username, cty.StringVal(terraformConfig.VsphereCredentials.Username))
	vsphereCredBlockBody.SetAttributeValue(vsphere.Vcenter, cty.StringVal(terraformConfig.VsphereCredentials.Vcenter))
	vsphereCredBlockBody.SetAttributeValue(vsphere.VcenterPort, cty.StringVal(terraformConfig.VsphereCredentials.VcenterPort))
}
