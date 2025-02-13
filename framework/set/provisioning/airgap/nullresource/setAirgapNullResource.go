package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
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
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, description})
	nullResourceBlockBody := nullResourceBlock.Body()

	if len(dependsOn) > 0 {
		var dependsOnValue hclwrite.Tokens
		for _, dep := range dependsOn {
			dependsOnValue = hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(dep)},
			}
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnValue)
	}

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	var bastionHostExpression string

	if terraformConfig.Module == modules.ImportK3s {
		bastionHostExpression = `"${` + defaults.AwsInstance + `.` + k3sServerOne + `.` + defaults.PublicIp + `}"`
	} else if terraformConfig.Module == modules.ImportRKE2 {
		bastionHostExpression = `"${` + defaults.AwsInstance + `.` + rke2ServerOne + `.` + defaults.PublicIp + `}"`
	} else {
		bastionHostExpression = defaults.AwsInstance + `.` + bastion + `.` + defaults.PublicIp
	}

	bastionHost := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(bastionHostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, bastionHost)

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	return provisionerBlockBody, nil
}
