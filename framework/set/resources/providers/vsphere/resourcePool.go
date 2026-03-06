package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereResourcePool is a function that will set the vSphere resource pool data source in the main.tf file.
func CreateVsphereResourcePool(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, dataCenterValue hclwrite.Tokens) {
	resourcePoolBlock := rootBody.AppendNewBlock(general.Data, []string{vsphere.VsphereResourcePool, vsphere.VsphereResourcePool})
	resourcePoolBlockBody := resourcePoolBlock.Body()

	resourcePoolBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.VsphereConfig.Pool))
	resourcePoolBlockBody.SetAttributeRaw(datacenterID, dataCenterValue)
}
