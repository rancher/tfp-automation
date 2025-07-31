package rke2k3s

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	v2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	"github.com/zclconf/go-cty/cty"
)

// SetRancher2ClusterV2 is a function that will set the rancher2_cluster_v2 configurations in the main.tf file.
func SetRancher2ClusterV2(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) error {
	rancher2ClusterV2Block := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.ClusterV2, terraformConfig.ResourcePrefix})
	rancher2ClusterV2BlockBody := rancher2ClusterV2Block.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Rancher2 + "." + defaults.StandardUser)},
	}

	rancher2ClusterV2BlockBody.SetAttributeRaw(defaults.Provider, provider)
	rancher2ClusterV2BlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	rancher2ClusterV2BlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(terratestConfig.KubernetesVersion))

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		v2.SetProxyConfig(rancher2ClusterV2BlockBody, terraformConfig)
	}

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
			v2.CreateRegistrySecret(terraformConfig, rootBody)
		}

		v2.SetMachineSelectorConfig(rkeConfigBlockBody, terraformConfig)

		registryBlock := rkeConfigBlockBody.AppendNewBlock(defaults.PrivateRegistries, nil)
		registryBlockBody := registryBlock.Body()

		v2.SetPrivateRegistryConfig(registryBlockBody, terraformConfig)
	}

	if strings.Contains(terraformConfig.Module, clustertypes.CUSTOM) && strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) {
		dependsOnBlock := `[` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows]`

		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnBlock)},
		}

		rancher2ClusterV2BlockBody.SetAttributeRaw(defaults.DependsOn, server)
	}

	return nil
}
