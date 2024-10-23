package instances

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetAwsInstances is a function that will set the AWS instances configurations in the main.tf file.
func SetAwsInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, clusterConfig *config.TerratestConfig) error {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, defaults.AwsInstance})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(clusterConfig.NodeCount))
	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.AWSConfig.AMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	awsSecGroupsList := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, awsSecGroupsList)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.AWSConfig.AWSKeyName))

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()
	rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s-${`+defaults.Count+`.`+defaults.Index+`}"`, terraformConfig.HostnamePrefix)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(defaults.Name, tags)

	return nil
}
