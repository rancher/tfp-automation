package locals

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	customnodepools "github.com/rancher/tfp-automation/framework/set/provisioning/custom/nodepools"
	"github.com/zclconf/go-cty/cty"
)

const (
	allPublicIPs = "all_public_ips"
	noProxy      = "localhost,127.0.0.0/8,10.0.0.0/8,172.0.0.0/8,192.168.0.0/16,.svc,.cluster.local,cattle-system.svc,169.254.169.25"
)

// SetLocals is a function that will set the locals configurations in the main.tf file.
func SetLocals(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	configMap []map[string]any, newFile *hclwrite.File, file *os.File, customClusterNames []string) (*os.File, error) {
	localsBlock := rootBody.AppendNewBlock(general.Locals, nil)
	localsBlockBody := localsBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		if terraformConfig.DownstreamClusterProvider == aws.Aws {
			expression, err := customnodepools.BuildAWSPublicIPExpression(terraformConfig, terratestConfig)
			if err != nil {
				return nil, err
			}

			value := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
			}

			localsBlockBody.SetAttributeRaw(allPublicIPs, value)
		}
	}

	roleFlags, err := customnodepools.BuildRoleFlags(terraformConfig, terratestConfig)
	if err != nil {
		return nil, err
	}

	roleFlagValues := make([]cty.Value, 0, len(roleFlags))
	for _, roleFlag := range roleFlags {
		roleFlagValues = append(roleFlagValues, cty.StringVal(roleFlag))
	}

	localsBlockBody.SetAttributeValue(clusters.RoleFlags, cty.ListVal(roleFlagValues))

	totalNodeCount := customnodepools.TotalNodeCount(terratestConfig)
	resourcePrefixExpression := fmt.Sprintf(`[for i in range(%d) : "%s-${i}"]`, totalNodeCount, terraformConfig.ResourcePrefix)
	resourcePrefixValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(resourcePrefixExpression)},
	}

	localsBlockBody.SetAttributeRaw(clusters.ResourcePrefix, resourcePrefixValue)

	setV2ClusterLocalBlock(localsBlockBody, terraformConfig, customClusterNames)

	return file, nil
}

func setV2ClusterLocalBlock(localsBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, customClusterNames []string) {
	for _, name := range customClusterNames {
		setCustomClusterLocalBlock(localsBlockBody, name, terraformConfig)
	}

	//Temporary workaround until fetching insecure node command is available for rancher2_cluster_v2 resoureces with tfp-rancher2
	if strings.Contains(terraformConfig.Module, general.Custom) || strings.Contains(terraformConfig.Module, general.Airgap) {
		setCustomClusterLocalBlock(localsBlockBody, terraformConfig.ResourcePrefix, terraformConfig)
	}
}

func setCustomClusterLocalBlock(localsBlockBody *hclwrite.Body, name string, terraformConfig *config.TerraformConfig) {
	originalNodeCommandExpression := rancher2.ClusterV2 + "." + name + "." + clusters.ClusterRegistrationToken + "[0]." + clusters.NodeCommand
	originalNodeCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(originalNodeCommandExpression)},
	}

	localsBlockBody.SetAttributeRaw(name+"_"+clusters.OriginalNodeCommand, originalNodeCommand)

	windowsOriginalNodeCommandExpression := rancher2.ClusterV2 + "." + name + "." + clusters.ClusterRegistrationToken + "[0]." + clusters.WindowsNodeCommand
	windowsOriginalNodeCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(windowsOriginalNodeCommandExpression)},
	}

	localsBlockBody.SetAttributeRaw(name+"_"+clusters.WindowsOriginalNodeCommand, windowsOriginalNodeCommand)

	insecureNodeCommandExpression := fmt.Sprintf(`"${replace(local.%s_original_node_command, "curl", "curl --insecure")}"`, name)
	insecureNodeCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(insecureNodeCommandExpression)},
	}

	localsBlockBody.SetAttributeRaw(name+"_"+clusters.InsecureNodeCommand, insecureNodeCommand)

	windowsInsecureNodeCommandExpression := fmt.Sprintf(`"${replace(local.%s_windows_original_node_command, "curl.exe", "curl.exe --insecure")}"`, name)
	windowsInsecureNodeCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(windowsInsecureNodeCommandExpression)},
	}

	localsBlockBody.SetAttributeRaw(name+"_"+clusters.InsecureWindowsNodeCommand, windowsInsecureNodeCommand)
}
