package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	serverOne   = "server1"
	serverTwo   = "server2"
	serverThree = "server3"
)

// CreateImportedNullResource is a helper function that will create the null_resource for the cluster.
func CreateImportedNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, publicDNS, resourceName string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, resourceName})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	hostExpression := `"` + publicDNS + `"`

	newHost := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, newHost)

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))

	if terraformConfig.Provider == defaults.Aws {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	} else if terraformConfig.Provider == defaults.Vsphere {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.VsphereConfig.VsphereUser))
	}

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)

	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne
	serverTwoName := terraformConfig.ResourcePrefix + `_` + serverTwo
	serverThreeName := terraformConfig.ResourcePrefix + `_` + serverThree

	var dependsOnServer string
	if terraformConfig.Provider == defaults.Aws {
		dependsOnServer = `[` + defaults.AwsInstance + `.` + serverOneName + `, ` + defaults.AwsInstance + `.` + serverTwoName + `, ` + defaults.AwsInstance + `.` + serverThreeName + `]`
	} else if terraformConfig.Provider == defaults.Vsphere {
		dependsOnServer = `[` + defaults.VsphereVirtualMachine + `.` + serverOneName + `, ` + defaults.VsphereVirtualMachine + `.` + serverTwoName + `, ` + defaults.VsphereVirtualMachine + `.` + serverThreeName + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	rootBody.AppendNewline()

	return nullResourceBlockBody, provisionerBlockBody
}
