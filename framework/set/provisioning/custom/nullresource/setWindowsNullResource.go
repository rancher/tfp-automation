package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetWindowsNullResource is a function that will set the Windows null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func SetWindowsNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, defaults.RegisterNodes + "-" + terraformConfig.ResourcePrefix + "-windows"})
	nullResourceBlockBody := nullResourceBlock.Body()

	countExpression := defaults.Length + `(` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows)`
	nullResourceBlockBody.SetAttributeRaw(defaults.Count, hclwrite.TokensForIdentifier(countExpression))

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))

	hostExpression := defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows[` + defaults.Count + `.` + defaults.Index + `].` + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)
	connectionBlockBody.SetAttributeValue(defaults.TargetPlatform, cty.StringVal(defaults.Windows))

	keyPathExpression := defaults.File + `("` + terraformConfig.WindowsPrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)

	regCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`["powershell.exe ${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" + defaults.InsecureWindowsNodeCommand + `}"]`)},
	}

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)

	return nil
}
