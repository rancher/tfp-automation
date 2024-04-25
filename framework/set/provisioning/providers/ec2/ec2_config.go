package ec2

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/amazon"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/zclconf/go-cty/cty"
)

const (
	region = "region"
)

// SetEC2RKE2K3SMachineConfig is a helper function that will set the EC2 RKE2/K3S
// Terraform machine configurations in the main.tf file.
func SetEC2RKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	ec2ConfigBlock := machineConfigBlockBody.AppendNewBlock(amazon.EC2Config, nil)
	ec2ConfigBlockBody := ec2ConfigBlock.Body()

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
