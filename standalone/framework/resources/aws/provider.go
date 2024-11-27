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
	kubeConfigPath    = "~/.kube/config"
	locals            = "locals"
	requiredProviders = "required_providers"
	rke2ServerOne     = "rke2_server1"
	rke2ServerTwo     = "rke2_server2"
	rke2ServerThree   = "rke2_server3"
)

// createTerraformProviderBlock will up the terraform block with the required aws provider.
func createTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	awsProviderVersion := os.Getenv("AWS_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.AwsSource),
		defaults.Version: cty.StringVal(awsProviderVersion),
	}))
}

// createAWSProviderBlock will set up the aws provider block.
func createAWSProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	awsProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Aws})
	awsProvBlockBody := awsProvBlock.Body()

	awsProvBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsProvBlockBody.SetAttributeValue(defaults.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsProvBlockBody.SetAttributeValue(defaults.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))
}

// createLocalBlock will set up the local block. Returns the local block.
func createLocalBlock(rootBody *hclwrite.Body) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	instanceIds := map[string]interface{}{
		rke2ServerOne:   defaults.AwsInstance + "." + rke2ServerOne + ".id",
		rke2ServerTwo:   defaults.AwsInstance + "." + rke2ServerTwo + ".id",
		rke2ServerThree: defaults.AwsInstance + "." + rke2ServerThree + ".id",
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
