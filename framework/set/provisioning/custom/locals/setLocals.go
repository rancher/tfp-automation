package locals

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetLocals is a function that will set the locals configurations in the main.tf file.
func SetLocals(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, clusterName string) error {
	localsBlock := rootBody.AppendNewBlock(defaults.Locals, nil)
	localsBlockBody := localsBlock.Body()

	localsBlockBody.SetAttributeValue(defaults.RoleFlags, cty.ListVal([]cty.Value{
		cty.StringVal(defaults.EtcdRoleFlag),
		cty.StringVal(defaults.ControlPlaneRoleFlag),
		cty.StringVal(defaults.WorkerRoleFlag),
	}))

	// Temporary workaround until fetching insecure node command is available for rancher2_cluster_v2 resoureces with tfp-rancher2
	if terraformConfig.Module == modules.CustomEC2RKE2 || terraformConfig.Module == modules.CustomEC2K3s {
		originalNodeCommandExpressionClusterV2 := defaults.ClusterV2 + "." + clusterName + "." + defaults.ClusterRegistrationToken + "[0]." + defaults.NodeCommand
		originalNodeCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(originalNodeCommandExpressionClusterV2)},
		}

		localsBlockBody.SetAttributeRaw(defaults.OriginalNodeCommand, originalNodeCommand)

		insecureNodeCommandExpressionClusterV2 := `"curl --insecure ${substr(local.original_node_command, 4, length(local.original_node_command) - 4)}"`
		insecureNodeCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(insecureNodeCommandExpressionClusterV2)},
		}

		localsBlockBody.SetAttributeRaw(defaults.InsecureNodeCommand, insecureNodeCommand)
	}

	return nil
}
