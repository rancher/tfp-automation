package rke2k3s

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	v2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	"github.com/zclconf/go-cty/cty"
)

// SetRancher2ClusterV2 is a function that will set the rancher2_cluster_v2 configurations in the main.tf file.
func SetRancher2ClusterV2(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, clusterName string) error {
	rancher2ClusterV2Block := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.ClusterV2, clusterName})
	rancher2ClusterV2BlockBody := rancher2ClusterV2Block.Body()

	rancher2ClusterV2BlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))
	rancher2ClusterV2BlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(terratestConfig.KubernetesVersion))

	rkeConfigBlock := rancher2ClusterV2BlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	if strings.Contains(terraformConfig.Module, "rke2") {
		machineGlobalConfigValue := hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOF\ncni: " + terraformConfig.CNI + "\nEOF"},
		})
		rkeConfigBlockBody.SetAttributeRaw(defaults.MachineGlobalConfig, machineGlobalConfigValue)
	}

	if terraformConfig.PrivateRegistries != nil {
		if terraformConfig.PrivateRegistries.Username != "" {
			rootBody.AppendNewline()
			v2.CreateRegistrySecret(terraformConfig, clusterName, rootBody)
		}

		v2.SetMachineSelectorConfig(rkeConfigBlockBody, terraformConfig)

		registryBlock := rkeConfigBlockBody.AppendNewBlock(defaults.PrivateRegistries, nil)
		registryBlockBody := registryBlock.Body()

		v2.SetPrivateRegistryConfig(registryBlockBody, terraformConfig)
	}

	return nil
}
