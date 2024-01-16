package ec2

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/zclconf/go-cty/cty"
)

// SetEC2RKE1Provider is a helper function that will set the EC2 RKE1 Terraform configurations in the main.tf file.
func SetEC2RKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	ec2ConfigBlock := nodeTemplateBlockBody.AppendNewBlock("amazonec2_config", nil)
	ec2ConfigBlockBody := ec2ConfigBlock.Body()

	ec2ConfigBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2ConfigBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
	ec2ConfigBlockBody.SetAttributeValue("ami", cty.StringVal(terraformConfig.AWSConfig.AMI))
	ec2ConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.AWSConfig.Region))
	awsSecGroupNames := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroupNames)
	ec2ConfigBlockBody.SetAttributeRaw("security_group", awsSecGroupNames)
	ec2ConfigBlockBody.SetAttributeValue("subnet_id", cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	ec2ConfigBlockBody.SetAttributeValue("vpc_id", cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	ec2ConfigBlockBody.SetAttributeValue("zone", cty.StringVal(terraformConfig.AWSConfig.AWSZoneLetter))
	ec2ConfigBlockBody.SetAttributeValue("root_size", cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))
	ec2ConfigBlockBody.SetAttributeValue("instance_type", cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
}

// SetEC2RKE2K3SProvider is a helper function that will set the EC2 RKE2/K3S Terraform provider details in the main.tf file.
func SetEC2RKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	ec2CredBlock := cloudCredBlockBody.AppendNewBlock("amazonec2_credential_config", nil)
	ec2CredBlockBody := ec2CredBlock.Body()

	ec2CredBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2CredBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
}
