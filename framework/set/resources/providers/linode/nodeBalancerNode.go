package linode

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	linodeDefalts "github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
	"github.com/zclconf/go-cty/cty"
)

// CreateNodeBalancerNode is a function that will set the node balancer configurations in the main.tf file.
func CreateNodeBalancerNode(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	nodeBalancerNodeBlock := rootBody.AppendNewBlock(general.Resource, []string{linodeDefalts.LinodeNodeBalancerNode, linodeDefalts.LinodeNodeBalancerNode + "_" + strconv.FormatInt(port, 10)})
	nodeBalancerNodeBlockBody := nodeBalancerNodeBlock.Body()

	expression := `{
        for instance in [` + linodeDefalts.LinodeInstance + `.` + serverOne + `, ` +
		linodeDefalts.LinodeInstance + `.` + serverTwo + `, ` +
		linodeDefalts.LinodeInstance + `.` + serverThree + `] : instance.label => instance
	}`

	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nodeBalancerNodeBlockBody.SetAttributeRaw(forEach, values)

	expression = linodeDefalts.LinodeNodeBalancer + `.` + linodeDefalts.LinodeNodeBalancer + `.id`
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nodeBalancerNodeBlockBody.SetAttributeRaw(linode.NodeBalancerID, values)

	expression = linodeDefalts.LinodeNodeBalancerConfig + `.` + linodeDefalts.LinodeNodeBalancerConfig + `_` + strconv.FormatInt(port, 10) + `.id`
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nodeBalancerNodeBlockBody.SetAttributeRaw(linode.ConfigID, values)

	expression = `each.key`
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nodeBalancerNodeBlockBody.SetAttributeRaw(linode.Label, values)

	expression = ` "${each.value.private_ip_address}:` + strconv.FormatInt(port, 10) + `"`
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(expression)},
	}

	nodeBalancerNodeBlockBody.SetAttributeRaw(linode.Address, values)
	nodeBalancerNodeBlockBody.SetAttributeValue(linode.Mode, cty.StringVal(accept))
}
