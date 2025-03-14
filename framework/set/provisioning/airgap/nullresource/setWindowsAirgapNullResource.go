package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetWindowsAirgapNullResource is a function that will set the Windows airgap null_resource configurations in the main.tf
// file, to register the nodes to the cluster
func SetWindowsAirgapNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, description string,
	dependsOn []string) (*hclwrite.Body, error) {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, description})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	bastionHostExpression := defaults.AwsInstance + `.` + bastion + `.` + defaults.PublicIp

	bastionHost := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(bastionHostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, bastionHost)

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))

	keyPathExpression := defaults.File + `("` + terraformConfig.WindowsPrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	return provisionerBlockBody, nil
}
