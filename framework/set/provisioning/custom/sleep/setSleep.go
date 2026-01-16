package sleep

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/zclconf/go-cty/cty"
)

// SetTimeSleep is a function that will set the time_sleep configurations in the main.tf file,
func SetTimeSleep(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, duration, dependsOnValue, suffix string) error {
	sleepResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.TimeSleep, general.TimeSleep + "-" + terraformConfig.ResourcePrefix + "-" + suffix})
	sleepResourceBlockBody := sleepResourceBlock.Body()

	sleepResourceBlockBody.SetAttributeValue(general.CreateDuration, cty.StringVal(duration))

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnValue)},
	}

	sleepResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

	return nil
}
