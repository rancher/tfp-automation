package rke2k3s

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	awsDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	vsphereDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/vsphere"
)

// getProviderIPAddresses is a helper function that returns the IP addresses of the nodes
func getProviderIPAddresses(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, rootBody *hclwrite.Body,
	serverOneName string) (string, string, string, string) {
	var nodeOnePublicIP, nodeOnePrivateIP, nodeTwoPublicIP, nodeThreePublicIP string

	serverTwoName := terraformConfig.ResourcePrefix + `_` + serverTwo
	serverThreeName := terraformConfig.ResourcePrefix + `_` + serverThree

	instances := []string{serverOneName, serverTwoName, serverThreeName}

	if terraformConfig.Provider == vsphereDefaults.Vsphere {
		dataCenterExpression := fmt.Sprintf(general.Data + `.` + vsphereDefaults.VsphereDatacenter + `.` + vsphereDefaults.VsphereDatacenter + `.id`)
		dataCenterValue := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dataCenterExpression)},
		}

		vsphere.CreateVsphereDatacenter(rootBody, terraformConfig)
		rootBody.AppendNewline()

		vsphere.CreateVsphereComputeCluster(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()

		vsphere.CreateVsphereNetwork(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()

		vsphere.CreateVsphereDatastore(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()

		vsphere.CreateVsphereVirtualMachineTemplate(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()
	}

	for _, instance := range instances {
		switch terraformConfig.Provider {
		case awsDefaults.Aws:
			aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()

			nodeOnePrivateIP = fmt.Sprintf("${%s.%s.private_ip}", awsDefaults.AwsInstance, serverOneName)
			nodeOnePublicIP = fmt.Sprintf("${%s.%s.public_ip}", awsDefaults.AwsInstance, serverOneName)
			nodeTwoPublicIP = fmt.Sprintf("${%s.%s.public_ip}", awsDefaults.AwsInstance, serverTwoName)
			nodeThreePublicIP = fmt.Sprintf("${%s.%s.public_ip}", awsDefaults.AwsInstance, serverThreeName)
		case vsphereDefaults.Vsphere:
			vsphere.CreateVsphereVirtualMachine(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()

			nodeOnePrivateIP = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, serverOneName)
			nodeOnePublicIP = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, serverOneName)
			nodeTwoPublicIP = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, serverTwoName)
			nodeThreePublicIP = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, serverThreeName)
		}
	}

	return nodeOnePublicIP, nodeOnePrivateIP, nodeTwoPublicIP, nodeThreePublicIP
}
