package aws

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/format"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
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
	configBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.AwsInstance, hostnamePrefix + "-windows"})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(general.Count, cty.NumberIntVal(terratestConfig.WindowsNodeCount))

	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) && strings.Contains(terraformConfig.Module, "2019") {
		configBlockBody.SetAttributeValue(aws.Ami, cty.StringVal(terraformConfig.AWSConfig.Windows2019AMI))
	} else if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) && strings.Contains(terraformConfig.Module, "2022") {
		configBlockBody.SetAttributeValue(aws.Ami, cty.StringVal(terraformConfig.AWSConfig.Windows2022AMI))
	}

	configBlockBody.SetAttributeValue(aws.InstanceType, cty.StringVal(terraformConfig.AWSConfig.WindowsInstanceType))
	configBlockBody.SetAttributeValue(aws.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(aws.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(aws.KeyName, cty.StringVal(terraformConfig.AWSConfig.WindowsKeyName))

	configBlockBody.AppendNewline()

	rootBlockDevice := configBlockBody.AppendNewBlock(aws.RootBlockDevice, nil)
	rootBlockDeviceBody := rootBlockDevice.Body()

	rootBlockDeviceBody.SetAttributeValue(aws.VolumeSize, cty.NumberIntVal(terraformConfig.AWSConfig.AWSRootSize))

	configBlockBody.AppendNewline()

	tagsBlock := configBlockBody.AppendNewBlock(general.Tags+" =", nil)
	tagsBlockBody := tagsBlock.Body()

	expression := fmt.Sprintf(`"%s-windows-${`+general.Count+`.`+general.Index+`}"`, terraformConfig.ResourcePrefix)
	tags := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	tagsBlockBody.SetAttributeRaw(aws.Name, tags)

	configBlockBody.AppendNewline()

	configBlockBody.SetAttributeValue(general.UserData, cty.StringVal(userData))
	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.WinRM))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))
	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) && strings.Contains(terraformConfig.Module, "2019") {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2019Password))
	} else if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) && strings.Contains(terraformConfig.Module, "2022") {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2022Password))
	}

	connectionBlockBody.SetAttributeValue(general.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(general.UseNTLM, cty.BoolVal(true))
	hostExpression := general.Self + "." + general.PublicIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	connectionBlockBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))
	configBlockBody.AppendNewline()

	provisionerBlock := configBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo Connected!!!"),
	}))

	if terraformConfig.Module == modules.ImportEC2RKE2Windows2019 || terraformConfig.Module == modules.ImportEC2RKE2Windows2022 {
		serverTwoName := terraformConfig.ResourcePrefix + `_server2`
		serverThreeName := terraformConfig.ResourcePrefix + `_server3`
		dependsOnServer := `[` + general.NullResource + `.` + serverTwoName + `, ` + general.NullResource + `.` + serverThreeName + `]`

		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		configBlockBody.AppendNewline()
		configBlockBody.SetAttributeRaw(general.DependsOn, server)
	}
}

// CreateAirgappedWindowsAWSInstances is a function that will set the Windows AWS instances configurations in the main.tf file.
func CreateAirgappedWindowsAWSInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	configBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.AwsInstance, hostnamePrefix})
	configBlockBody := configBlock.Body()

	configBlockBody.SetAttributeValue(aws.AssociatePublicIPAddress, cty.BoolVal(false))

	if strings.Contains(terraformConfig.Module, modules.AirgapRKE2Windows2019) {
		configBlockBody.SetAttributeValue(aws.Ami, cty.StringVal(terraformConfig.AWSConfig.Windows2019AMI))
	} else if strings.Contains(terraformConfig.Module, modules.AirgapRKE2Windows2022) {
		configBlockBody.SetAttributeValue(aws.Ami, cty.StringVal(terraformConfig.AWSConfig.Windows2022AMI))
	}

	configBlockBody.SetAttributeValue(aws.InstanceType, cty.StringVal(terraformConfig.AWSConfig.WindowsInstanceType))
	configBlockBody.SetAttributeValue(aws.SubnetId, cty.StringVal(terraformConfig.AWSConfig.AWSSubnetID))
	securityGroups := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	configBlockBody.SetAttributeRaw(aws.VpcSecurityGroupIds, securityGroups)
	configBlockBody.SetAttributeValue(aws.KeyName, cty.StringVal(terraformConfig.AWSConfig.WindowsKeyName))

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

	configBlockBody.SetAttributeValue(general.UserData, cty.StringVal(userData))
	configBlockBody.AppendNewline()

	connectionBlock := configBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.WinRM))
	connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.WindowsAWSUser))
	if strings.Contains(terraformConfig.Module, modules.AirgapRKE2Windows2019) {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2019Password))
	} else if strings.Contains(terraformConfig.Module, modules.AirgapRKE2Windows2022) {
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.AWSConfig.Windows2022Password))
	}

	connectionBlockBody.SetAttributeValue(general.Insecure, cty.BoolVal(true))
	connectionBlockBody.SetAttributeValue(general.UseNTLM, cty.BoolVal(true))

	hostExpression := general.Self + "." + general.PrivateIp
	host := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(general.Host, host)

	connectionBlockBody.SetAttributeValue(aws.Timeout, cty.StringVal(terraformConfig.AWSConfig.Timeout))
}
