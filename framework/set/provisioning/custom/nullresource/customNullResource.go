package nullresource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CustomNullResource is a function that will set the null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func CustomNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) error {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, defaults.RegisterNodes + "-" + terraformConfig.ResourcePrefix})
	nullResourceBlockBody := nullResourceBlock.Body()

	var countExpression string
	if strings.Contains(terraformConfig.Provider, defaults.Aws) {
		countExpression = defaults.Length + `(` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `)`
	} else if strings.Contains(terraformConfig.Provider, defaults.Vsphere) {
		countExpression = defaults.Length + `(` + defaults.VsphereVirtualMachine + `.` + terraformConfig.ResourcePrefix + `)`
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.Count, hclwrite.TokensForIdentifier(countExpression))

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	if strings.Contains(terraformConfig.Module, defaults.Custom) && strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		regCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + defaults.Cluster + `.` + terraformConfig.ResourcePrefix + `.` +
				defaults.ClusterRegistrationToken + `[0].` + defaults.NodeCommand + `} ${` + defaults.Local + `.` + defaults.RoleFlags +
				`[` + defaults.Count + `.` + defaults.Index + `]} ` + defaults.NodeNameFlag + ` ${` + defaults.Local + `.` +
				defaults.ResourcePrefix + `[` + defaults.Count + `.` + defaults.Index + `]}"]`)},
		}

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)
	}

	if strings.Contains(terraformConfig.Module, defaults.Custom) && !strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		regCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" +
				defaults.InsecureNodeCommand + `} ${` + defaults.Local + `.` + defaults.RoleFlags + `[` + defaults.Count + `.` +
				defaults.Index + `]} ` + defaults.NodeNameFlag + ` ${` + defaults.Local + `.` +
				defaults.ResourcePrefix + `[` + defaults.Count + `.` + defaults.Index + `]}"]`)},
		}

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)
	}

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))

	var hostExpression string

	if terraformConfig.Provider == defaults.Aws {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
		hostExpression = fmt.Sprintf(`"${%s.%s[%s.%s].%s}"`, defaults.AwsInstance, terraformConfig.ResourcePrefix, defaults.Count, defaults.Index, defaults.PublicIp)
	} else if terraformConfig.Provider == defaults.Vsphere {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.VsphereConfig.VsphereUser))
		hostExpression = fmt.Sprintf(`"${%s.%s[%s.%s].%s}"`, defaults.VsphereVirtualMachine, terraformConfig.ResourcePrefix, defaults.Count, defaults.Index, defaults.DefaultIPAddress)
	}

	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)

	if strings.Contains(terraformConfig.Module, defaults.Custom) && !strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		regCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" +
				defaults.InsecureNodeCommand + `} ${` + defaults.Local + `.` + defaults.RoleFlags + `[` + defaults.Count + `.` +
				defaults.Index + `]} ` + defaults.NodeNameFlag + ` ${` + defaults.Local + `.` +
				defaults.ResourcePrefix + `[` + defaults.Count + `.` + defaults.Index + `]}"]`)},
		}

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)
	}

	if strings.Contains(terraformConfig.Module, defaults.Custom) && strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		clusterExpression := `[` + defaults.Cluster + `.` + terraformConfig.ResourcePrefix + `]`
		cluster := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(clusterExpression)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, cluster)
	}

	if strings.Contains(terraformConfig.Module, defaults.Custom) && !strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		clusterV2Expression := `[` + defaults.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
		clusterV2 := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(clusterV2Expression)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, clusterV2)
	}

	return nil
}
