package locals

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
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
		expression := fmt.Sprintf(`concat(`+aws.AwsInstance+`.%s-etcd.*.public_ip, `+
			aws.AwsInstance+`.%s-control-plane.*.public_ip, `+
			aws.AwsInstance+`.%s-worker.*.public_ip)`, terraformConfig.ResourcePrefix, terraformConfig.ResourcePrefix, terraformConfig.ResourcePrefix)
		value := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		localsBlockBody.SetAttributeRaw(allPublicIPs, value)

		expression = fmt.Sprintf(`concat([for i in range(%d) : "--etcd"], [for i in range(%d) : "--controlplane"], [for i in range(%d) : "--worker"])`,
			terratestConfig.EtcdCount, terratestConfig.ControlPlaneCount, terratestConfig.WorkerCount)
		value = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		localsBlockBody.SetAttributeRaw(clusters.RoleFlags, value)
	} else {
		var roleFlags []cty.Value

		for range terratestConfig.EtcdCount {
			roleFlags = append(roleFlags, cty.StringVal(clusters.EtcdRoleFlag))
		}

		for range terratestConfig.ControlPlaneCount {
			roleFlags = append(roleFlags, cty.StringVal(clusters.ControlPlaneRoleFlag))
		}

		for range terratestConfig.WorkerCount {
			roleFlags = append(roleFlags, cty.StringVal(clusters.WorkerRoleFlag))
		}

		localsBlockBody.SetAttributeValue(clusters.RoleFlags, cty.ListVal(roleFlags))
	}

	totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
	resourcePrefixExpression := fmt.Sprintf(`[for i in range(%d) : "%s-${i}"]`, totalNodeCount, terraformConfig.ResourcePrefix)
	resourcePrefixValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(resourcePrefixExpression)},
	}

	localsBlockBody.SetAttributeRaw(clusters.ResourcePrefix, resourcePrefixValue)

	if !strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		setV2ClusterLocalBlock(localsBlockBody, terraformConfig, customClusterNames)
	}

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

	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) && (terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "") {
		setWindowsProxyLocalBlock(localsBlockBody, name)
	}
}

func setWindowsProxyLocalBlock(localsBlockBody *hclwrite.Body, name string) error {
	// Terraform, by design, results to a .cmd file. Need to explictily call powershell.exe
	envReplace := fmt.Sprintf(`replace(local.%s_windows_original_node_command, "$env:", "powershell.exe $env:")`, name)
	curlReplace := fmt.Sprintf(`"${replace(%s, "curl.exe", "curl.exe --insecure")}"`, envReplace)

	proxyWindowsInsecureNodeCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(curlReplace)},
	}

	localsBlockBody.SetAttributeRaw(name+"_"+clusters.InsecureWindowsProxyNodeCommand, proxyWindowsInsecureNodeCommand)

	return nil
}
