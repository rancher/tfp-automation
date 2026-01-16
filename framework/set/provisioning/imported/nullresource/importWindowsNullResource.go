package nullresource

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/zclconf/go-cty/cty"
)

// CreateImportedWindowsNullResource is a helper function that will create the null_resource for the Windows node.
func CreateImportedWindowsNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	publicDNS, resourceName string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, resourceName})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	var hostExpression string

	if terratestConfig.WindowsNodeCount == 1 {
		hostExpression = `"${` + aws.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows[0].` + general.PublicIp + `}"`
	} else if terratestConfig.WindowsNodeCount > 1 {
		hostExpression = `"${` + aws.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows[` + general.Count + `.` + general.Index + `].` + general.PublicIp + `}"`
	}

	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.WinRM))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))

	if strings.Contains(terraformConfig.Module, modules.ImportEC2RKE2Windows2019) {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2019Password))
	} else if strings.Contains(terraformConfig.Module, modules.ImportEC2RKE2Windows2022) {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2022Password))
	}

	connectionBlockBody.SetAttributeValue(general.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(general.UseNTLM, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	return nullResourceBlockBody, provisionerBlockBody
}
