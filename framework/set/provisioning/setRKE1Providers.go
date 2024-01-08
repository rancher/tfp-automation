package provisioning

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/zclconf/go-cty/cty"
)

// setEC2RKE1 is a helper function that will set the EC2 RKE1 Terraform configurations in the main.tf file.
func setEC2RKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	ec2ConfigBlock := nodeTemplateBlockBody.AppendNewBlock("amazonec2_config", nil)
	ec2ConfigBlockBody := ec2ConfigBlock.Body()

	ec2ConfigBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2ConfigBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
	ec2ConfigBlockBody.SetAttributeValue("ami", cty.StringVal(terraformConfig.AWSConfig.Ami))
	ec2ConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.AWSConfig.Region))
	awsSecGroupNames := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroupNames)
	ec2ConfigBlockBody.SetAttributeRaw("security_group", awsSecGroupNames)
	ec2ConfigBlockBody.SetAttributeValue("subnet_id", cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	ec2ConfigBlockBody.SetAttributeValue("vpc_id", cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	ec2ConfigBlockBody.SetAttributeValue("zone", cty.StringVal(terraformConfig.AWSConfig.AWSZoneLetter))
	ec2ConfigBlockBody.SetAttributeValue("root_size", cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))
	ec2ConfigBlockBody.SetAttributeValue("instance_type", cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
}

// setLinodeRKE1 is a helper function that will set the Linode RKE1 Terraform configurations in the main.tf file.
func setLinodeRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeConfigBlock := nodeTemplateBlockBody.AppendNewBlock("linode_config", nil)
	linodeConfigBlockBody := linodeConfigBlock.Body()

	linodeConfigBlockBody.SetAttributeValue("image", cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	linodeConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.LinodeConfig.Region))
	linodeConfigBlockBody.SetAttributeValue("root_pass", cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	linodeConfigBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeConfig.LinodeToken))
}
