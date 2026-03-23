package nullresource

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

const (
	allPublicIPs = "all_public_ips"
)

// CustomNullResource is a function that will set the null_resource configurations in the main.tf file,
// to register the nodes to the cluster
func CustomNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) error {
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, general.RegisterNodes + "-" + terraformConfig.ResourcePrefix})
	nullResourceBlockBody := nullResourceBlock.Body()

	var countExpression string
	if strings.Contains(terraformConfig.Provider, aws.Aws) {
		countExpression = general.Length + `(` + general.Local + `.` + allPublicIPs + `)`
	} else if strings.Contains(terraformConfig.Provider, vsphere.Vsphere) {
		countExpression = general.Length + `(` + vsphere.VsphereVirtualMachine + `.` + terraformConfig.ResourcePrefix + `)`
	}

	nullResourceBlockBody.SetAttributeRaw(general.Count, hclwrite.TokensForIdentifier(countExpression))

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	regCommand := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + general.Local + `.` + terraformConfig.ResourcePrefix + "_" +
			clusters.InsecureNodeCommand + `} ${` + general.Local + `.` + clusters.RoleFlags + `[` + general.Count + `.` +
			general.Index + `]} ` + clusters.NodeNameFlag + ` ${` + general.Local + `.` +
			clusters.ResourcePrefix + `[` + general.Count + `.` + general.Index + `]}"]`)},
	}

	provisionerBlockBody.SetAttributeRaw(general.Inline, regCommand)

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))

	var hostExpression string

	switch terraformConfig.Provider {
	case aws.Aws:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
		hostExpression = fmt.Sprintf(`%s.%s[%s.%s]`, general.Local, allPublicIPs, general.Count, general.Index)

	case vsphere.Vsphere:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.VsphereConfig.VsphereUser))
		hostExpression = fmt.Sprintf(`"${%s.%s[%s.%s].%s}"`, vsphere.VsphereVirtualMachine, terraformConfig.ResourcePrefix, general.Count, general.Index, general.DefaultIPAddress)
	}

	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)

	regCommand = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`["${` + general.Local + `.` + terraformConfig.ResourcePrefix + "_" +
			clusters.InsecureNodeCommand + `} ${` + general.Local + `.` + clusters.RoleFlags + `[` + general.Count + `.` +
			general.Index + `]} ` + clusters.NodeNameFlag + ` ${` + general.Local + `.` +
			clusters.ResourcePrefix + `[` + general.Count + `.` + general.Index + `]}"]`)},
	}

	provisionerBlockBody.SetAttributeRaw(general.Inline, regCommand)

	clusterV2Expression := `[` + rancher2.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
	clusterV2 := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(clusterV2Expression)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, clusterV2)

	return nil
}
