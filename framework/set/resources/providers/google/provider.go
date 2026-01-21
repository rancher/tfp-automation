package google

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
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

// CreateGoogleCloudTerraformProviderBlock will up the terraform block with the required Google Cloud provider.
func CreateGoogleCloudTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(googleDefaults.Google, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(googleDefaults.GoogleSource),
		general.Version: cty.StringVal(cloudProviderVersion),
	}))
}

// CreateGoogleCloudProviderBlock will set up the Google Cloud provider block.
func CreateGoogleCloudProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	googleProvBlock := rootBody.AppendNewBlock(general.Provider, []string{googleDefaults.Google})
	googleProvBlockBody := googleProvBlock.Body()

	googleProvBlockBody.SetAttributeValue(googleDefaults.GoogleProject, cty.StringVal(terraformConfig.GoogleConfig.ProjectID))
	googleProvBlockBody.SetAttributeValue(googleDefaults.GoogleRegion, cty.StringVal(terraformConfig.GoogleConfig.Region))
	googleProvBlockBody.SetAttributeValue(googleDefaults.GoogleZone, cty.StringVal(terraformConfig.GoogleConfig.Zone))
}

// CreateGoogleCloudLocalBlock will set up the local block. Returns the local block.
func CreateGoogleCloudLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	instanceIds := map[string]any{
		serverOne:   googleDefaults.GoogleComputeInstance + "." + serverOne + ".id",
		serverTwo:   googleDefaults.GoogleComputeInstance + "." + serverTwo + ".id",
		serverThree: googleDefaults.GoogleComputeInstance + "." + serverThree + ".id",
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
