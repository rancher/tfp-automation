package linode

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	linodeDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
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

	reqProvsBlockBody.SetAttributeValue(linodeDefaults.Linode, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(linodeDefaults.LinodeSource),
		general.Version: cty.StringVal(cloudProviderVersion),
	}))
}

// CreateLinodeProviderBlock will set up the linode provider block.
func CreateLinodeProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeProvBlock := rootBody.AppendNewBlock(general.Provider, []string{linodeDefaults.Linode})
	linodeProvBlockBody := linodeProvBlock.Body()

	linodeProvBlockBody.SetAttributeValue(linode.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))
}

// CreateLinodeLocalBlock will set up the local block. Returns the local block.
func CreateLinodeLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	instanceIds := map[string]any{
		serverOne:   linodeDefaults.LinodeInstance + "." + serverOne + ".id",
		serverTwo:   linodeDefaults.LinodeInstance + "." + serverTwo + ".id",
		serverThree: linodeDefaults.LinodeInstance + "." + serverThree + ".id",
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
