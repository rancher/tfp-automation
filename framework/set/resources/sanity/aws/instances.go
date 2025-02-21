package aws

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateAWSInstances is a function that will set the AWS instances configurations in the main.tf file.
func CreateAWSInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, "custom") {
		configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(terratestConfig.NodeCount))
	}

	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.AWSConfig.AMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))

	var awsSecGroupsExpression string

	if strings.Contains(terraformConfig.Module, "custom") {
		awsSecGroupsExpression = fmt.Sprintf(`["%s"]`, terraformConfig.AWSConfig.AWSSecurityGroupNames[0])
	} else {
		awsSecGroupsExpression = fmt.Sprintf(`["%s"]`, terraformConfig.AWSConfig.StandaloneSecurityGroupNames[0])
	}

	awsSecGroupsList := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(awsSecGroupsExpression)},
	}

	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, awsSecGroupsList)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.AWSConfig.AWSKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()

	if strings.Contains(hostnamePrefix, "registry") {
		rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.RegistryRootSize))
	} else {
		rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))
	}

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	if strings.Contains(terraformConfig.Module, "custom") {
		expression := fmt.Sprintf(`"%s-${`+defaults.Count+`.`+defaults.Index+`}"`, terraformConfig.HostnamePrefix)
		tags := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		tagsBlockBody.SetAttributeRaw(defaults.Name, tags)
	} else {
		expression := fmt.Sprintf(`"%s`, terraformConfig.HostnamePrefix+"-"+hostnamePrefix+`"`)
		tags := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		tagsBlockBody.SetAttributeRaw(defaults.Name, tags)
	}

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

	hostExpression := defaults.Self + "." + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("cloud-init status --wait"),
	}))
}
