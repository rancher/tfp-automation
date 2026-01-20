package aws

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/zclconf/go-cty/cty"
)

const (
	ecrRegistry = "ecr_registry"
)

// CreateAWSInstances is a function that will set the AWS instances configurations in the main.tf file.
func CreateAWSInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
		configBlockBody.SetAttributeValue(general.Count, cty.NumberIntVal(totalNodeCount))
	}

	configBlockBody.SetAttributeValue(aws.Ami, cty.StringVal(terraformConfig.AWSConfig.AMI))

	configBlockBody.SetAttributeValue(aws.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
	configBlockBody.SetAttributeValue(aws.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(aws.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(aws.KeyName, cty.StringVal(terraformConfig.AWSConfig.AWSKeyName))

	configBlockBody.AppendNewline()

	if terraformConfig.AWSConfig.IPAddressType == aws.IPv6 {
		configBlockBody.SetAttributeValue(aws.EnablePrimaryIPv6, cty.BoolVal(true))
		configBlockBody.SetAttributeValue(aws.IPV6AddressCount, cty.NumberIntVal(1))

		metadataOptionsBlock := configBlockBody.AppendNewBlock(metadataOptions, nil)
		metadataOptionsBlockBody := metadataOptionsBlock.Body()

		metadataOptionsBlockBody.SetAttributeValue(httpProtocolIPv6, cty.StringVal(terraformConfig.AWSConfig.HTTPProtocolIPv6))
		configBlockBody.AppendNewline()
	}

	rootBlockDevice := configBlockBody.AppendNewBlock(aws.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()

	if strings.Contains(hostnamePrefix, general.Registry) {
		rootBlockDeviceBody.SetAttributeValue(aws.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.RegistryRootSize))
	} else {
		rootBlockDeviceBody.SetAttributeValue(aws.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))
	}

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(general.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		expression := fmt.Sprintf(`"%s-${`+general.Count+`.`+general.Index+`}"`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix)
		tags := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		tagsBlockBody.SetAttributeRaw(aws.Name, tags)
	} else {
		expression := fmt.Sprintf(`"%s`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix+`"`)
		tags := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		tagsBlockBody.SetAttributeRaw(aws.Name, tags)
	}

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	hostExpression := general.Self + "." + general.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))

	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))
}

// CreateAirgappedAWSInstances is a function that will set the AWS instances configurations in the main.tf file.
func CreateAirgappedAWSInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(aws.AssociatePublicIPAddress, cty.BoolVal(false))
	configBlockBody.SetAttributeValue(aws.Ami, cty.StringVal(terraformConfig.AWSConfig.AMI))
	configBlockBody.SetAttributeValue(aws.InstanceType, cty.StringVal(terraformConfig.AWSConfig.AWSInstanceType))
	configBlockBody.SetAttributeValue(aws.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(aws.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(aws.KeyName, cty.StringVal(terraformConfig.AWSConfig.AWSKeyName))
	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(aws.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()
	rootBlockDeviceBody.SetAttributeValue(aws.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(general.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix+`"`)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(aws.Name, tags)

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))
	var hostExpression string
	if terraformConfig.AWSConfig.IPAddressType == aws.IPv6 {
		hostExpression = general.Self + "." + general.IPV6Addresses + `[0]`
	} else {
		hostExpression = general.Self + "." + general.PrivateIp
	}

	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	connectionBlockBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))
}
