package linode

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	locals            = "locals"
	requiredProviders = "required_providers"
	instanceIDs       = "instance_ids"
	serverOne         = "server1"
	serverTwo         = "server2"
	serverThree       = "server3"
)

// CreateLinodeTerraformProviderBlock will up the terraform block with the required linode provider.
func CreateLinodeTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Linode, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.LinodeSource),
		defaults.Version: cty.StringVal(cloudProviderVersion),
	}))
}

// CreateLinodeProviderBlock will set up the linode provider block.
func CreateLinodeProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Linode})
	linodeProvBlockBody := linodeProvBlock.Body()

	linodeProvBlockBody.SetAttributeValue(linode.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))
}

// CreateLinodeLocalBlock will set up the local block. Returns the local block.
func CreateLinodeLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	instanceIds := map[string]any{
		serverOne:   defaults.LinodeInstance + "." + serverOne + ".id",
		serverTwo:   defaults.LinodeInstance + "." + serverTwo + ".id",
		serverThree: defaults.LinodeInstance + "." + serverThree + ".id",
	}

	instanceIdsBlock := localBlockBody.AppendNewBlock(instanceIDs+" =", nil)
	instanceIdsBlockBody := instanceIdsBlock.Body()

	for key, value := range instanceIds {
		expression := value.(string)
		instanceValues := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		instanceIdsBlockBody.SetAttributeRaw(key, instanceValues)
	}
}
