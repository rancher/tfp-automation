package imported

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetImportedCluster is a function that will set the imported rancher2_cluster configurations in the main.tf file.
func SetImportedCluster(rootBody *hclwrite.Body, clusterName string) error {
	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, clusterName})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(defaults.Description, cty.StringVal("tfp-automation imported cluster"))

	return nil
}
