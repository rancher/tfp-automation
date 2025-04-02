package linode

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	accept        = "accept"
	algorithm     = "algorithm"
	check         = "check"
	checkAttempts = "check_attempts"
	checkTimeout  = "check_timeout"
	connection    = "connection"
	forEach       = "for_each"
	none          = "none"
	protocol      = "protocol"
	roundRobin    = "roundrobin"
	stickiness    = "stickiness"
	tcp           = "tcp"
)

// CreateNodeBalancerConfig is a function that will set the node balancer configurations in the main.tf file.
func CreateNodeBalancerConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, port int64) {
	nodeBalancerConfigBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LinodeNodeBalancerConfig, defaults.LinodeNodeBalancerConfig + "_" + strconv.FormatInt(port, 10)})
	nodeBalancerConfigBlockBody := nodeBalancerConfigBlock.Body()

	expression := defaults.LinodeNodeBalancer + `.` + defaults.LinodeNodeBalancer + `.id`
	nodeBalancerID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	nodeBalancerConfigBlockBody.SetAttributeRaw(linode.NodeBalancerID, nodeBalancerID)
	nodeBalancerConfigBlockBody.SetAttributeValue(defaults.Port, cty.NumberIntVal(port))
	nodeBalancerConfigBlockBody.SetAttributeValue(protocol, cty.StringVal(tcp))
	nodeBalancerConfigBlockBody.SetAttributeValue(check, cty.StringVal(connection))
	nodeBalancerConfigBlockBody.SetAttributeValue(checkAttempts, cty.NumberIntVal(3))
	nodeBalancerConfigBlockBody.SetAttributeValue(checkTimeout, cty.NumberIntVal(30))
	nodeBalancerConfigBlockBody.SetAttributeValue(stickiness, cty.StringVal(none))
	nodeBalancerConfigBlockBody.SetAttributeValue(algorithm, cty.StringVal(roundRobin))
}
