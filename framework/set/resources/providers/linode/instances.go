package linode

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateLinodeInstances is a function that will set the Linode instances configurations in the main.tf file.
func CreateLinodeInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LinodeInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, defaults.Custom) {
		totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
		configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(totalNodeCount))
	}

	configBlockBody.SetAttributeValue(linode.Image, cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	configBlockBody.SetAttributeValue(linode.Region, cty.StringVal(terraformConfig.LinodeConfig.Region))
	configBlockBody.SetAttributeValue(linode.Type, cty.StringVal(terraformConfig.LinodeConfig.Type))
	configBlockBody.SetAttributeValue(linode.RootPass, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	configBlockBody.SetAttributeValue(linode.SwapSize, cty.NumberIntVal(terraformConfig.LinodeConfig.SwapSize))
	configBlockBody.SetAttributeValue(linode.PrivateIP, cty.BoolVal(terraformConfig.LinodeConfig.PrivateIP))
	configBlockBody.SetAttributeValue(linode.Label, cty.StringVal(terraformConfig.ResourcePrefix+"-"+hostnamePrefix))

	tags := format.ListOfStrings(terraformConfig.LinodeConfig.Tags)
	configBlockBody.SetAttributeRaw(linode.Tags, tags)

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(linode.RootUser))
	connectionBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))

	hostExpression := defaults.Self + "." + defaults.IPAddress
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.LinodeConfig.Timeout))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))
}
