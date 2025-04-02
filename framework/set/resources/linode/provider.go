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
	k3sServerOne      = "k3s_server1"
	k3sServerTwo      = "k3s_server2"
	k3sServerThree    = "k3s_server3"
	requiredProviders = "required_providers"
	rke2InstanceIDs   = "rke2_instance_ids"
	rke2ServerOne     = "rke2_server1"
	rke2ServerTwo     = "rke2_server2"
	rke2ServerThree   = "rke2_server3"
)

// CreateLinodeTerraformProviderBlock will up the terraform block with the required linode provider.
func CreateLinodeTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	linodeProviderVersion := os.Getenv("LINODE_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Linode, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.LinodeSource),
		defaults.Version: cty.StringVal(linodeProviderVersion),
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

	var instanceIds map[string]interface{}
	if terraformConfig.Standalone.RKE2Version != "" {
		instanceIds = map[string]interface{}{
			rke2ServerOne:   defaults.LinodeInstance + "." + rke2ServerOne + ".id",
			rke2ServerTwo:   defaults.LinodeInstance + "." + rke2ServerTwo + ".id",
			rke2ServerThree: defaults.LinodeInstance + "." + rke2ServerThree + ".id",
		}
	} else if terraformConfig.Standalone.K3SVersion != "" {
		instanceIds = map[string]interface{}{
			k3sServerOne:   defaults.LinodeInstance + "." + k3sServerOne + ".id",
			k3sServerTwo:   defaults.LinodeInstance + "." + k3sServerTwo + ".id",
			k3sServerThree: defaults.LinodeInstance + "." + k3sServerThree + ".id",
		}
	}

	instanceIdsBlock := localBlockBody.AppendNewBlock(rke2InstanceIDs+" =", nil)
	instanceIdsBlockBody := instanceIdsBlock.Body()

	for key, value := range instanceIds {
		expression := value.(string)
		instanceValues := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		instanceIdsBlockBody.SetAttributeRaw(key, instanceValues)
	}
}
