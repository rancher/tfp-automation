package vsphere

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereComputeCluster is a function that will set the vSphere compute cluster configuration in the main.tf file.
func CreateVsphereComputeCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, dataCenterValue hclwrite.Tokens) {
	computeClusterBlock := rootBody.AppendNewBlock(general.Data, []string{vsphere.VsphereComputeCluster, vsphere.VsphereComputeCluster})
	computeClusterBlockBody := computeClusterBlock.Body()

	computeClusterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.VsphereConfig.HostSystem))
	computeClusterBlockBody.SetAttributeRaw(datacenterID, dataCenterValue)
}
