package locals

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetLocals is a function that will set the locals configurations in the main.tf file.
func SetLocals(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, configMap []map[string]any, newFile *hclwrite.File, file *os.File, customClusterNames []string) (*os.File, error) {
	localsBlock := rootBody.AppendNewBlock(defaults.Locals, nil)
	localsBlockBody := localsBlock.Body()

	localsBlockBody.SetAttributeValue(defaults.RoleFlags, cty.ListVal([]cty.Value{
		cty.StringVal(defaults.EtcdRoleFlag),
		cty.StringVal(defaults.ControlPlaneRoleFlag),
		cty.StringVal(defaults.WorkerRoleFlag),
	}))

	if customClusterNames != nil {
		for _, name := range customClusterNames {
			// Temporary workaround until fetching insecure node command is available for rancher2_cluster_v2 resoureces with tfp-rancher2
			originalNodeCommandExpressionClusterV2 := defaults.ClusterV2 + "." + name + "." + defaults.ClusterRegistrationToken + "[0]." + defaults.NodeCommand
			originalNodeCommand := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(originalNodeCommandExpressionClusterV2)},
			}

			localsBlockBody.SetAttributeRaw(name+"_"+defaults.OriginalNodeCommand, originalNodeCommand)

			insecureNodeCommandExpressionClusterV2 := fmt.Sprintf(`"curl --insecure ${substr(local.%s_original_node_command, 4, length(local.%s_original_node_command) - 4)}"`, name, name)
			insecureNodeCommand := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(insecureNodeCommandExpressionClusterV2)},
			}

			localsBlockBody.SetAttributeRaw(name+"_"+defaults.InsecureNodeCommand, insecureNodeCommand)

		}
	} else {
		//Temporary workaround until fetching insecure node command is available for rancher2_cluster_v2 resoureces with tfp-rancher2
		if terraformConfig.Module == modules.CustomEC2RKE2 || terraformConfig.Module == modules.CustomEC2K3s ||
			terraformConfig.Module == modules.AirgapRKE2 || terraformConfig.Module == modules.AirgapK3S {
			originalNodeCommandExpressionClusterV2 := defaults.ClusterV2 + "." + terraformConfig.ResourcePrefix + "." + defaults.ClusterRegistrationToken + "[0]." + defaults.NodeCommand
			originalNodeCommand := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(originalNodeCommandExpressionClusterV2)},
			}

			localsBlockBody.SetAttributeRaw(terraformConfig.ResourcePrefix+"_"+defaults.OriginalNodeCommand, originalNodeCommand)

			insecureNodeCommandExpressionClusterV2 := fmt.Sprintf(`"curl --insecure ${substr(local.%s_original_node_command, 4, length(local.%s_original_node_command) - 4)}"`, terraformConfig.ResourcePrefix, terraformConfig.ResourcePrefix)
			insecureNodeCommand := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(insecureNodeCommandExpressionClusterV2)},
			}

			localsBlockBody.SetAttributeRaw(terraformConfig.ResourcePrefix+"_"+defaults.InsecureNodeCommand, insecureNodeCommand)
		}
	}

	return file, nil
}
