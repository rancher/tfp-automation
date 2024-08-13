package aws

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/amazon"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/zclconf/go-cty/cty"
)

const (
	cloudCredential = "rancher2_cloud_credential"

	accessKey    = "access_key"
	secretKey    = "secret_key"
	region       = "region"
	resource     = "resource"
	resourceName = "name"
)

// SetAWSRKE1Provider is a helper function that will set the AWS RKE1
// Terraform configurations in the main.tf file.
func SetAWSRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	awsConfigBlock := nodeTemplateBlockBody.AppendNewBlock(amazon.EC2Config, nil)
	awsConfigBlockBody := awsConfigBlock.Body()

	awsConfigBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsConfigBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))

	awsConfigBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsConfigBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))
	awsConfigBlockBody.SetAttributeValue(amazon.AMI, cty.StringVal(terraformConfig.AWSConfig.AMI))
	awsConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsSecGroupNames := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroupNames)
	awsConfigBlockBody.SetAttributeRaw(amazon.SecurityGroup, awsSecGroupNames)
	awsConfigBlockBody.SetAttributeValue(amazon.SubnetID, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	awsConfigBlockBody.SetAttributeValue(amazon.VPCID, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	awsConfigBlockBody.SetAttributeValue(amazon.Zone, cty.StringVal(terraformConfig.AWSConfig.AWSZoneLetter))
	awsConfigBlockBody.SetAttributeValue(amazon.RootSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))
	awsConfigBlockBody.SetAttributeValue(amazon.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
}

// SetAWSRKE2K3SProvider is a helper function that will set the AWS RKE2/K3S
// Terraform provider details in the main.tf file.
func SetAWSRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(resource, []string{cloudCredential, cloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.CloudCredentialName))

	awsCredBlock := cloudCredBlockBody.AppendNewBlock(amazon.EC2CredentialConfig, nil)
	awsCredBlockBody := awsCredBlock.Body()

	awsCredBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
	awsCredBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))
}
