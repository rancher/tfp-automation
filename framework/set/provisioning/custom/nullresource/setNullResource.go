package nullresource

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetNullResource is a function that will set the null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func SetNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, defaults.RegisterNodes + "-" + terraformConfig.ResourcePrefix})
	nullResourceBlockBody := nullResourceBlock.Body()

	countExpression := defaults.Length + `(` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `)`
	nullResourceBlockBody.SetAttributeRaw(defaults.Count, hclwrite.TokensForIdentifier(countExpression))

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	if terraformConfig.Module == modules.CustomEC2RKE1 {
		regCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + defaults.Cluster + `.` + terraformConfig.ResourcePrefix + `.` + defaults.ClusterRegistrationToken + `[0].` + defaults.NodeCommand + `} ${` + defaults.Local + `.` + defaults.RoleFlags + `[` + defaults.Count + `.` + defaults.Index + `]}"]`)},
		}

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)
	}

	if terraformConfig.Module == modules.CustomEC2RKE2 || terraformConfig.Module == modules.CustomEC2K3s ||
		terraformConfig.Module == modules.CustomEC2RKE2Windows {
		regCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" + defaults.InsecureNodeCommand + `} ${` + defaults.Local + `.` + defaults.RoleFlags + `[` + defaults.Count + `.` + defaults.Index + `]}"]`)},
		}

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)
	}

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

	hostExpression := defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `[` + defaults.Count + `.` + defaults.Index + `].` + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)

	if terraformConfig.Module == modules.CustomEC2RKE2 || terraformConfig.Module == modules.CustomEC2K3s ||
		terraformConfig.Module == modules.CustomEC2RKE2Windows {
		regCommand := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + defaults.Local + `.` + terraformConfig.ResourcePrefix + "_" + defaults.InsecureNodeCommand + `} ${` + defaults.Local + `.` + defaults.RoleFlags + `[` + defaults.Count + `.` + defaults.Index + `]}"]`)},
		}

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, regCommand)
	}

	if terraformConfig.Module == modules.CustomEC2RKE1 {
		clusterExpression := `[` + defaults.Cluster + `.` + terraformConfig.ResourcePrefix + `]`
		cluster := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(clusterExpression)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, cluster)
	}

	if terraformConfig.Module == modules.CustomEC2RKE2 || terraformConfig.Module == modules.CustomEC2K3s ||
		terraformConfig.Module == modules.CustomEC2RKE2Windows {
		clusterV2Expression := `[` + defaults.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
		clusterV2 := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(clusterV2Expression)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, clusterV2)
	}

	return nil
}
