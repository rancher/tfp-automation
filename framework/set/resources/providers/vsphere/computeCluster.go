package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereComputeCluster is a function that will set the vSphere compute cluster configuration in the main.tf file.
func CreateVsphereComputeCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, dataCenterValue hclwrite.Tokens) {
	computeClusterBlock := rootBody.AppendNewBlock(defaults.Data, []string{defaults.VsphereComputeCluster, defaults.VsphereComputeCluster})
	computeClusterBlockBody := computeClusterBlock.Body()

	computeClusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.VsphereConfig.HostSystem))
	computeClusterBlockBody.SetAttributeRaw(datacenterID, dataCenterValue)
}
