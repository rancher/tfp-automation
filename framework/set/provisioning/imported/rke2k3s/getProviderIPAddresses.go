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
	linuxNodeNames []string) (map[string]string, map[string]string) {
	nodePublicIPs := make(map[string]string, len(linuxNodeNames))
	nodePrivateIPs := make(map[string]string, len(linuxNodeNames))

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

	for _, instance := range linuxNodeNames {
		switch terraformConfig.Provider {
		case awsDefaults.Aws:
			aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()

			nodePrivateIPs[instance] = fmt.Sprintf("${%s.%s.private_ip}", awsDefaults.AwsInstance, instance)
			nodePublicIPs[instance] = fmt.Sprintf("${%s.%s.public_ip}", awsDefaults.AwsInstance, instance)
		case vsphereDefaults.Vsphere:
			vsphere.CreateVsphereVirtualMachine(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()

			nodePrivateIPs[instance] = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, instance)
			nodePublicIPs[instance] = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, instance)
		}
	}

	return nodePublicIPs, nodePrivateIPs
}
