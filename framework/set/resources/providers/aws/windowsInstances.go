package aws

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	userData = `<<-EOF
              <powershell>
              winrm quickconfig -q
              winrm set winrm/config/service '@{AllowUnencrypted="true"}'
              winrm set winrm/config/service/auth '@{Basic="true"}'

              netsh advfirewall firewall add rule name="WinRM HTTP" dir=in action=allow protocol=TCP localport=5985
              netsh advfirewall firewall add rule name="WinRM HTTPS" dir=in action=allow protocol=TCP localport=5986
              </powershell>
              EOF`
)

// CreateWindowsAWSInstances is a function that will set the Windows AWS instances configurations in the main.tf file.
func CreateWindowsAWSInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, hostnamePrefix + "-windows"})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(terratestConfig.WindowsNodeCount))

	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.AWSConfig.WindowsAMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.AWSConfig.WindowsInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.AWSConfig.WindowsKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()

	rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s-windows-${`+defaults.Count+`.`+defaults.Index+`}"`, terraformConfig.ResourcePrefix)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(defaults.Name, tags)

	configBlockBody.AppendNewline()

	configBlockBody.SetAttributeValue(defaults.UserData, cty.StringVal(userData))

	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.WinRM))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))
	connectionBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSPassword))
	connectionBlockBody.SetAttributeValue(defaults.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(defaults.UseNTLM, cty.BoolVal(true))

	hostExpression := defaults.Self + "." + defaults.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))
	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))

	if terraformConfig.Module == modules.ImportEC2RKE2Windows {
		serverTwoName := terraformConfig.ResourcePrefix + `_server2`
		serverThreeName := terraformConfig.ResourcePrefix + `_server3`
		dependsOnServer := `[` + defaults.NullResource + `.` + serverTwoName + `, ` + defaults.NullResource + `.` + serverThreeName + `]`

		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		configBlockBody.AppendNewline()
		configBlockBody.SetAttributeRaw(defaults.DependsOn, server)
	}
}

// CreateAirgappedWindowsAWSInstances is a function that will set the Windows AWS instances configurations in the main.tf file.
func CreateAirgappedWindowsAWSInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(defaults.AssociatePublicIPAddress, cty.BoolVal(false))
	configBlockBody.SetAttributeValue(defaults.Ami, cty.StringVal(terraformConfig.AWSConfig.WindowsAMI))
	configBlockBody.SetAttributeValue(defaults.InstanceType, cty.StringVal(terraformConfig.AWSConfig.WindowsInstanceType))
	configBlockBody.SetAttributeValue(defaults.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))

	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(defaults.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(defaults.KeyName, cty.StringVal(terraformConfig.AWSConfig.WindowsKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(defaults.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()
	rootBlockDeviceBody.SetAttributeValue(defaults.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(defaults.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s`, terraformConfig.ResourcePrefix+"-"+hostnamePrefix+`"`)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(defaults.Name, tags)

	configBlockBody.AppendNewline()

	configBlockBody.SetAttributeValue(defaults.UserData, cty.StringVal(userData))
	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.WinRM))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))
	connectionBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSPassword))
	connectionBlockBody.SetAttributeValue(defaults.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(defaults.UseNTLM, cty.BoolVal(true))

	hostExpression := defaults.Self + "." + defaults.PrivateIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, host)

	connectionBlockBody.SetAttributeValue(defaults.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))
}
