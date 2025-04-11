package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateImportedWindowsNullResource is a helper function that will create the null_resource for the Windows node.
func CreateImportedWindowsNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	publicDNS, resourceName string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, resourceName})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	var hostExpression string

	if terratestConfig.WindowsNodeCount == 1 {
		hostExpression = `"${` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows[0].` + defaults.PublicIp + `}"`
	} else if terratestConfig.WindowsNodeCount > 1 {
		hostExpression = `"${` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows[` + defaults.Count + `.` + defaults.Index + `].` + defaults.PublicIp + `}"`
	}

	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.WinRM))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))
	connectionBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSPassword))
	connectionBlockBody.SetAttributeValue(defaults.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(defaults.UseNTLM, cty.BoolVal(true))

	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	return nullResourceBlockBody, provisionerBlockBody
}
