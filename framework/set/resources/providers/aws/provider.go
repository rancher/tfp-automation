package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	locals            = "locals"
	requiredProviders = "required_providers"
	serverOne         = "server1"
	serverTwo         = "server2"
	serverThree       = "server3"
)

// CreateAWSTerraformProviderBlock will up the terraform block with the required aws provider.
func CreateAWSTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.AwsSource),
		defaults.Version: cty.StringVal(cloudProviderVersion),
	}))
}

// CreateAWSProviderBlock will set up the aws provider block.
func CreateAWSProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	awsProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Aws})
	awsProvBlockBody := awsProvBlock.Body()

	awsProvBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsProvBlockBody.SetAttributeValue(defaults.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsProvBlockBody.SetAttributeValue(defaults.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))
}

// CreateAWSLocalBlock will set up the local block. Returns the local block.
func CreateAWSLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	var instanceIds map[string]any
	if !terraformConfig.AWSConfig.EnablePrimaryIPv6 {
		instanceIds = map[string]any{
			serverOne:   defaults.AwsInstance + "." + serverOne + ".id",
			serverTwo:   defaults.AwsInstance + "." + serverTwo + ".id",
			serverThree: defaults.AwsInstance + "." + serverThree + ".id",
		}
	} else if terraformConfig.AWSConfig.EnablePrimaryIPv6 {
		instanceIds = map[string]any{
			serverOne:   defaults.AwsInstance + "." + serverOne + ".ipv6_addresses[0]",
			serverTwo:   defaults.AwsInstance + "." + serverTwo + ".ipv6_addresses[0]",
			serverThree: defaults.AwsInstance + "." + serverThree + ".ipv6_addresses[0]",
		}
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
