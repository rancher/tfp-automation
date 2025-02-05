package rke1

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetRancher2Cluster is a function that will set the rancher2_cluster configurations in the main.tf file.
func SetRancher2Cluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, clusterName string) error {
	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, clusterName})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	networkBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(defaults.Plugin, cty.StringVal(terraformConfig.CNI))

	return nil
}
