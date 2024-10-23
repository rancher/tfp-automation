package rke2k3s

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	cniPlaceholder = "placeholder" // Placeholder value -- to be removed
)

// setRancher2ClusterV2 is a function that will set the rancher2_cluster_v2 configurations in the main.tf file.
func setRancher2ClusterV2(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, clusterConfig *config.TerratestConfig, clusterName string) error {
	rancher2ClusterV2Block := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.ClusterV2, clusterName})
	rancher2ClusterV2BlockBody := rancher2ClusterV2Block.Body()

	rancher2ClusterV2BlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))
	rancher2ClusterV2BlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(clusterConfig.KubernetesVersion))

	rkeConfigBlock := rancher2ClusterV2BlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	machineGlobalConfigValue := hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\ncni: " + terraformConfig.CNI + "\nEOF"},
	})
	rkeConfigBlockBody.SetAttributeRaw(defaults.MachineGlobalConfig, machineGlobalConfigValue)

	return nil
}
