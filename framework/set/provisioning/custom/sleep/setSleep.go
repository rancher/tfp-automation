package sleep

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetTimeSleep is a function that will set the time_sleep configurations in the main.tf file,
func SetTimeSleep(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, duration, dependsOnValue string) error {
	sleepResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.TimeSleep, defaults.TimeSleep + "-" + terraformConfig.ResourcePrefix})
	sleepResourceBlockBody := sleepResourceBlock.Body()

	sleepResourceBlockBody.SetAttributeValue(defaults.CreateDuration, cty.StringVal(duration))

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnValue)},
	}

	sleepResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	return nil
}
