package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

const (
	serverOne   = "server1"
	serverTwo   = "server2"
	serverThree = "server3"
)

// CreateImportedNullResource is a helper function that will create the null_resource for the cluster.
func CreateImportedNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, publicDNS, resourceName string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, resourceName})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	hostExpression := `"` + publicDNS + `"`

	newHost := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, newHost)

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))

	switch terraformConfig.Provider {
	case aws.Aws:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	case vsphere.Vsphere:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.VsphereConfig.VsphereUser))
	}

	keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)

	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne
	serverTwoName := terraformConfig.ResourcePrefix + `_` + serverTwo
	serverThreeName := terraformConfig.ResourcePrefix + `_` + serverThree

	var dependsOnServer string
	switch terraformConfig.Provider {
	case aws.Aws:
		dependsOnServer = `[` + aws.AwsInstance + `.` + serverOneName + `, ` + aws.AwsInstance + `.` + serverTwoName + `, ` + aws.AwsInstance + `.` + serverThreeName + `]`
	case vsphere.Vsphere:
		dependsOnServer = `[` + vsphere.VsphereVirtualMachine + `.` + serverOneName + `, ` + vsphere.VsphereVirtualMachine + `.` + serverTwoName + `, ` + vsphere.VsphereVirtualMachine + `.` + serverThreeName + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

	rootBody.AppendNewline()

	return nullResourceBlockBody, provisionerBlockBody
}
