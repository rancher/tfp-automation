package aws

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/amazon"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/zclconf/go-cty/cty"
)

// SetAWSRKE2K3SMachineConfig is a helper function that will set the AWS RKE2/K3S
// Terraform machine configurations in the main.tf file.
func SetAWSRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, ami, instanceType string) {
	awsConfigBlock := machineConfigBlockBody.AppendNewBlock(amazon.EC2Config, nil)
	awsConfigBlockBody := awsConfigBlock.Body()

	awsConfigBlockBody.SetAttributeValue(aws.Region, cty.StringVal(terraformConfig.AWSConfig.Region))

	awsConfigBlockBody.SetAttributeValue(amazon.AMI, cty.StringVal(ami))
	awsConfigBlockBody.SetAttributeValue(amazon.InstanceType, cty.StringVal(instanceType))

	awsConfigBlockBody.SetAttributeValue(amazon.SSHUser, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	awsConfigBlockBody.SetAttributeValue(amazon.VolumeType, cty.StringVal(terraformConfig.AWSConfig.AWSVolumeType))
	awsConfigBlockBody.SetAttributeValue(amazon.RootSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroupNames)
	awsConfigBlockBody.SetAttributeRaw(amazon.SecurityGroup, securityGroups)

	awsConfigBlockBody.SetAttributeValue(amazon.SubnetID, cty.StringVal(terraformConfig.AWSConfig.RancherSubnetID))
	awsConfigBlockBody.SetAttributeValue(amazon.VPCID, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	awsConfigBlockBody.SetAttributeValue(amazon.Zone, cty.StringVal(terraformConfig.AWSConfig.AWSZoneLetter))

	if terraformConfig.AWSConfig.EnablePrimaryIPv6 {
		awsConfigBlockBody.SetAttributeValue(amazon.EnablePrimaryIPv6, cty.BoolVal(true))
		awsConfigBlockBody.SetAttributeValue(amazon.HTTPProtocolIPv6, cty.StringVal(terraformConfig.AWSConfig.HTTPProtocolIPv6))
		awsConfigBlockBody.SetAttributeValue(amazon.IPv6AddressCount, cty.StringVal(terraformConfig.AWSConfig.IPv6AddressCount))
		awsConfigBlockBody.SetAttributeValue(amazon.IPv6AddressOnly, cty.BoolVal(terraformConfig.AWSConfig.IPv6AddressOnly))
	}
}
