package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/zclconf/go-cty/cty"
)

const (
	alwaysRun     = "always_run"
	bastion       = "bastion"
	importNodeOne = "import_node_one"
	k3sServerOne  = "k3s_server1"
	rke2ServerOne = "rke2_server1"
)

// SetAirgapNullResource is a function that will set the airgap null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func SetAirgapNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, description string,
	dependsOn []string) (*hclwrite.Body, error) {
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, description})
	nullResourceBlockBody := nullResourceBlock.Body()

	if len(dependsOn) > 0 {
		var dependsOnValue hclwrite.Tokens
		for _, dep := range dependsOn {
			dependsOnValue = hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(dep)},
			}
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, dependsOnValue)
	}

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	var bastionHostExpression string

	switch terraformConfig.Module {
	case modules.ImportEC2K3s:
		bastionHostExpression = `"${` + aws.AwsInstance + `.` + k3sServerOne + `_` + terraformConfig.ResourcePrefix + `.` + general.PublicIp + `}"`
	case modules.ImportEC2RKE2:
		bastionHostExpression = `"${` + aws.AwsInstance + `.` + rke2ServerOne + `_` + terraformConfig.ResourcePrefix + `.` + general.PublicIp + `}"`
	default:
		bastionHostExpression = `"${` + aws.AwsInstance + `.` + bastion + `_` + terraformConfig.ResourcePrefix + `.` + general.PublicIp + `}"`
	}

	bastionHost := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(bastionHostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, bastionHost)

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

	keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	return provisionerBlockBody, nil
}
