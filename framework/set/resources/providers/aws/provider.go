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
	k3sServerOne      = "k3s_server1"
	k3sServerTwo      = "k3s_server2"
	k3sServerThree    = "k3s_server3"
	rke2ServerOne     = "rke2_server1"
	rke2ServerTwo     = "rke2_server2"
	rke2ServerThree   = "rke2_server3"
)

// CreateAWSTerraformProviderBlock will up the terraform block with the required aws provider.
func CreateAWSTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	awsProviderVersion := os.Getenv("AWS_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.AwsSource),
		defaults.Version: cty.StringVal(awsProviderVersion),
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

	var instanceIds map[string]interface{}
	if terraformConfig.Standalone.RKE2Version != "" {
		instanceIds = map[string]interface{}{
			rke2ServerOne:   defaults.AwsInstance + "." + rke2ServerOne + ".id",
			rke2ServerTwo:   defaults.AwsInstance + "." + rke2ServerTwo + ".id",
			rke2ServerThree: defaults.AwsInstance + "." + rke2ServerThree + ".id",
		}
	} else if terraformConfig.Standalone.K3SVersion != "" {
		instanceIds = map[string]interface{}{
			k3sServerOne:   defaults.AwsInstance + "." + k3sServerOne + ".id",
			k3sServerTwo:   defaults.AwsInstance + "." + k3sServerTwo + ".id",
			k3sServerThree: defaults.AwsInstance + "." + k3sServerThree + ".id",
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
