package rke1

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// SetRancher2Cluster is a function that will set the rancher2_cluster configurations in the main.tf file.
func SetRancher2Cluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) error {
	clusterBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.Cluster, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(clusters.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue(clusters.KubernetesVersion, cty.StringVal(terratestConfig.KubernetesVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock(clusters.Network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(clusters.Plugin, cty.StringVal(terraformConfig.CNI))

	return nil
}
