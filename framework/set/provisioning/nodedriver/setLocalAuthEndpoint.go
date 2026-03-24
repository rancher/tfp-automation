package nodedriver

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

const (
	enabled = "enabled"
)

// SetLocalAuthEndpoint is a function that will set the local auth endpoint configurations in the main.tf file.
func SetLocalAuthEndpoint(terraformConfig *config.TerraformConfig, rkeConfigBlockBody *hclwrite.Body) error {
	aceBlock := rkeConfigBlockBody.AppendNewBlock(clusters.LocalAuthEndpoint, nil)
	aceBlockBody := aceBlock.Body()

	aceBlockBody.SetAttributeValue(enabled, cty.BoolVal(terraformConfig.LocalAuthEndpoint))

	return nil
}
