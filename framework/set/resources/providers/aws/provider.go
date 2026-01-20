package aws

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
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

	reqProvsBlockBody.SetAttributeValue(aws.Aws, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(aws.AwsSource),
		general.Version: cty.StringVal(cloudProviderVersion),
	}))
}

// CreateAWSProviderBlock will set up the aws provider block.
func CreateAWSProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	awsProvBlock := rootBody.AppendNewBlock(general.Provider, []string{aws.Aws})
	awsProvBlockBody := awsProvBlock.Body()

	awsProvBlockBody.SetAttributeValue(aws.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsProvBlockBody.SetAttributeValue(aws.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsProvBlockBody.SetAttributeValue(aws.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))
}

// CreateAWSLocalBlock will set up the local block. Returns the local block.
func CreateAWSLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	var instanceIds map[string]any
	if terraformConfig.AWSConfig.IPAddressType != aws.IPv6 {
		instanceIds = map[string]any{
			serverOne:   aws.AwsInstance + "." + serverOne + ".id",
			serverTwo:   aws.AwsInstance + "." + serverTwo + ".id",
			serverThree: aws.AwsInstance + "." + serverThree + ".id",
		}
	} else if terraformConfig.AWSConfig.IPAddressType == aws.IPv6 {
		instanceIds = map[string]any{
			serverOne:   aws.AwsInstance + "." + serverOne + ".ipv6_addresses[0]",
			serverTwo:   aws.AwsInstance + "." + serverTwo + ".ipv6_addresses[0]",
			serverThree: aws.AwsInstance + "." + serverThree + ".ipv6_addresses[0]",
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
