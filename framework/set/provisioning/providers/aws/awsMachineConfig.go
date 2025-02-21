package aws

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/amazon"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetAWSRKE2K3SMachineConfig is a helper function that will set the AWS RKE2/K3S
// Terraform machine configurations in the main.tf file.
func SetAWSRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	awsConfigBlock := machineConfigBlockBody.AppendNewBlock(amazon.EC2Config, nil)
	awsConfigBlockBody := awsConfigBlock.Body()

	awsConfigBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
	awsConfigBlockBody.SetAttributeValue(amazon.AMI, cty.StringVal(terraformConfig.AWSConfig.AMI))
	awsConfigBlockBody.SetAttributeValue(amazon.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
	awsConfigBlockBody.SetAttributeValue(amazon.SSHUser, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	awsConfigBlockBody.SetAttributeValue(amazon.VolumeType, cty.StringVal(terraformConfig.AWSConfig.AWSVolumeType))
	awsConfigBlockBody.SetAttributeValue(amazon.RootSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	awsSecGroupsExpression := fmt.Sprintf(`["%s"]`, terraformConfig.AWSConfig.AWSSecurityGroupNames[0])
	awsSecGroupsList := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(awsSecGroupsExpression)},
	}
	awsConfigBlockBody.SetAttributeRaw(amazon.SecurityGroup, awsSecGroupsList)

	awsConfigBlockBody.SetAttributeValue(amazon.SubnetID, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	awsConfigBlockBody.SetAttributeValue(amazon.VPCID, cty.StringVal(terraformConfig.AWSConfig.AWSVpcID))
	awsConfigBlockBody.SetAttributeValue(amazon.Zone, cty.StringVal(terraformConfig.AWSConfig.AWSZoneLetter))
}
