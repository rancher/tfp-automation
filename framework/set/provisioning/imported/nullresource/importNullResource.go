package nullresource

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

// CreateImportedNullResource is a helper function that will create the null_resource for the cluster.
func CreateImportedNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, publicIP, resourceName string,
	linuxNodeNames []string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, resourceName})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	hostExpression := `"` + publicIP + `"`

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

	dependsOnResources := make([]string, 0, len(linuxNodeNames))
	switch terraformConfig.Provider {
	case aws.Aws:
		for _, nodeName := range linuxNodeNames {
			dependsOnResources = append(dependsOnResources, aws.AwsInstance+"."+nodeName)
		}
	case vsphere.Vsphere:
		for _, nodeName := range linuxNodeNames {
			dependsOnResources = append(dependsOnResources, vsphere.VsphereVirtualMachine+"."+nodeName)
		}
	}

	dependsOnServer := `[`
	if len(dependsOnResources) > 0 {
		dependsOnServer += strings.Join(dependsOnResources, ", ")
	}
	dependsOnServer += `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

	rootBody.AppendNewline()

	return nullResourceBlockBody, provisionerBlockBody
}
