package linode

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	linodeDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
	"github.com/zclconf/go-cty/cty"
)

// CreateLinodeInstances is a function that will set the Linode instances configurations in the main.tf file.
func CreateLinodeInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(general.Resource, []string{linodeDefaults.LinodeInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
		configBlockBody.SetAttributeValue(general.Count, cty.NumberIntVal(totalNodeCount))
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

	connectionBlock := configBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(linode.RootUser))
	connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))

	hostExpression := general.Self + "." + general.IPAddress
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	connectionBlockBody.SetAttributeValue(general.Timeout, cty.StringVal(terraformConfig.LinodeConfig.Timeout))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))
}
