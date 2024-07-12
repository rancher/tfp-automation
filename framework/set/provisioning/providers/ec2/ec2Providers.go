package ec2

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

// SetEC2RKE1Provider is a helper function that will set the EC2 RKE1
// Terraform configurations in the main.tf file.
func SetEC2RKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	ec2ConfigBlock := nodeTemplateBlockBody.AppendNewBlock(amazon.EC2Config, nil)
	ec2ConfigBlockBody := ec2ConfigBlock.Body()

	ec2ConfigBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2ConfigBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
	ec2ConfigBlockBody.SetAttributeValue(amazon.AMI, cty.StringVal(terraformConfig.AWSConfig.AMI))
	ec2ConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsSecGroupNames := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroupNames)
	ec2ConfigBlockBody.SetAttributeRaw(amazon.SecurityGroup, awsSecGroupNames)
	ec2ConfigBlockBody.SetAttributeValue(amazon.SubnetID, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	ec2ConfigBlockBody.SetAttributeValue(amazon.VPCID, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	ec2ConfigBlockBody.SetAttributeValue(amazon.Zone, cty.StringVal(terraformConfig.AWSConfig.AWSZoneLetter))
	ec2ConfigBlockBody.SetAttributeValue(amazon.RootSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))
	ec2ConfigBlockBody.SetAttributeValue(amazon.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
}

// SetEC2RKE2K3SProvider is a helper function that will set the EC2 RKE2/K3S
// Terraform provider details in the main.tf file.
func SetEC2RKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(resource, []string{cloudCredential, cloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.CloudCredentialName))

	ec2CredBlock := cloudCredBlockBody.AppendNewBlock(amazon.EC2CredentialConfig, nil)
	ec2CredBlockBody := ec2CredBlock.Body()

	ec2CredBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2CredBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
}
