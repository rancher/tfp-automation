package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CustomWindowsNullResource is a function that will set the Windows null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func CustomWindowsNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, defaults.RegisterNodes + "-" + terraformConfig.ResourcePrefix + "-windows"})
	nullResourceBlockBody := nullResourceBlock.Body()

	countExpression := defaults.Length + `(` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows)`
	nullResourceBlockBody.SetAttributeRaw(defaults.Count, hclwrite.TokensForIdentifier(countExpression))

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.WinRM))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))
	connectionBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSPassword))
	connectionBlockBody.SetAttributeValue(defaults.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(defaults.UseNTLM, cty.BoolVal(true))

	hostExpression := defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows[` + defaults.Count + `.` + defaults.Index + `].` + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	var regCommand hclwrite.Tokens

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		regCommand = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["powershell.exe ${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" + defaults.InsecureWindowsProxyNodeCommand + `}"]`)},
		}
	} else {
		regCommand = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["powershell.exe ${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" + defaults.InsecureWindowsNodeCommand + `}"]`)},
		}
	}

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)

	return nil
}
