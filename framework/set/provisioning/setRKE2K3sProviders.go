package provisioning

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/zclconf/go-cty/cty"
)

// setEC2RKE2K3S is a helper function that will set the EC2 RKE2/K3S Terraform provider details in the main.tf file.
func setEC2RKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	ec2CredBlock := cloudCredBlockBody.AppendNewBlock("amazonec2_credential_config", nil)
	ec2CredBlockBody := ec2CredBlock.Body()

	ec2CredBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2CredBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
}

// setLinodeRKE2K3S is a helper function that will set the Linode RKE2/K3S Terraform provider details in the main.tf file.
func setLinodeRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	linodeCredBlock := cloudCredBlockBody.AppendNewBlock("linode_credential_config", nil)
	linodeCredBlockBody := linodeCredBlock.Body()

	linodeCredBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeConfig.LinodeToken))
}

// setEC2RKE2K3SMachineConfig is a helper function that will set the EC2 RKE2/K3S Terraform machine configurations in the main.tf file.
func setEC2RKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	ec2ConfigBlock := machineConfigBlockBody.AppendNewBlock("amazonec2_config", nil)
	ec2ConfigBlockBody := ec2ConfigBlock.Body()

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

// setLinodeRKE2K3SMachineConfig is a helper function that will set the EC2 RKE2/K3S Terraform machine configurations in the main.tf file.
func setLinodeRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	linodeConfigBlock := machineConfigBlockBody.AppendNewBlock("linode_config", nil)
	linodeConfigBlockBody := linodeConfigBlock.Body()

	linodeConfigBlockBody.SetAttributeValue("image", cty.StringVal(terraformConfig.LinodeConfig.LinodeImage))
	linodeConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.LinodeConfig.Region))
	linodeConfigBlockBody.SetAttributeValue("root_pass", cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	linodeConfigBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeConfig.LinodeToken))
}
