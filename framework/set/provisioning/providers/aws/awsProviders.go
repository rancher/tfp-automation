package aws

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/amazon"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// SetAWSRKE2K3SProvider is a helper function that will set the AWS RKE2/K3S
// Terraform provider details in the main.tf file.
func SetAWSRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	awsCredBlock := cloudCredBlockBody.AppendNewBlock(amazon.EC2CredentialConfig, nil)
	awsCredBlockBody := awsCredBlock.Body()

	awsCredBlockBody.SetAttributeValue(aws.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsCredBlockBody.SetAttributeValue(aws.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))
}
