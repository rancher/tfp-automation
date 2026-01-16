package nullresource

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// CustomWindowsNullResource is a function that will set the Windows null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func CustomWindowsNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, clusterName string) error {
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, general.RegisterNodes + "-" + clusterName + "-windows"})
	nullResourceBlockBody := nullResourceBlock.Body()

	countExpression := general.Length + `(` + aws.AwsInstance + `.` + clusterName + `-windows)`
	nullResourceBlockBody.SetAttributeRaw(general.Count, hclwrite.TokensForIdentifier(countExpression))

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.WinRM))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))

	if strings.Contains(terraformConfig.Module, modules.CustomEC2RKE2Windows2019) {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2019Password))
	} else if strings.Contains(terraformConfig.Module, modules.CustomEC2RKE2Windows2022) {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2022Password))
	}

	connectionBlockBody.SetAttributeValue(general.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(general.UseNTLM, cty.BoolVal(true))

	hostExpression := aws.AwsInstance + `.` + clusterName + `-windows[` + general.Count + `.` + general.Index + `].` + general.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	var regCommand hclwrite.Tokens

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		regCommand = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["powershell.exe ${` + general.Local + `.` + clusterName + "_" + clusters.InsecureWindowsProxyNodeCommand + `}"]`)},
		}
	} else {
		regCommand = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["powershell.exe ${` + general.Local + `.` + clusterName + "_" + clusters.InsecureWindowsNodeCommand + `}"]`)},
		}
	}

	provisionerBlockBody.SetAttributeRaw(general.Inline, regCommand)
	return nil
}
