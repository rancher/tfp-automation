package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereDatacenter is a function that will set the vSphere data center configuration in the main.tf file.
func CreateVsphereDatacenter(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	datacenterBlock := rootBody.AppendNewBlock(general.Data, []string{vsphere.VsphereDatacenter, vsphere.VsphereDatacenter})
	datacenterBlockBody := datacenterBlock.Body()

	datacenterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.VsphereConfig.DataCenter))
}
