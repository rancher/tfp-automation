package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereDatastore is a function that will set the vSphere datastore configuration in the main.tf file.
func CreateVsphereDatastore(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, dataCenterValue hclwrite.Tokens) {
	datastoreBlock := rootBody.AppendNewBlock(defaults.Data, []string{defaults.VsphereDatastore, defaults.VsphereDatastore})
	datastoreBlockBody := datastoreBlock.Body()

	datastoreBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.VsphereConfig.DataStore))
	datastoreBlockBody.SetAttributeRaw(datacenterID, dataCenterValue)
}
