package rke2k3s

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	awsDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	vsphereDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/vsphere"
)

// SetCustomRKE2K3s is a function that will set the custom RKE2/K3s cluster configurations in the main.tf file.
func SetCustomRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	switch terraformConfig.Provider {
	case awsDefaults.Aws:
		aws.CreateCustomClusterAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix+"-etcd",
			terraformConfig.AWSConfig.AMI, terraformConfig.AWSConfig.AWSInstanceType, terratestConfig.EtcdCount)

		aws.CreateCustomClusterAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix+"-control-plane",
			terraformConfig.AWSConfig.AMI, terraformConfig.AWSConfig.AWSInstanceType, terratestConfig.ControlPlaneCount)

		if terraformConfig.MixedArchitecture {
			aws.CreateCustomClusterAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix+"-worker",
				terraformConfig.AWSConfig.ARMAMI, terraformConfig.AWSConfig.ARMInstanceType, terratestConfig.WorkerCount)
		} else {
			aws.CreateCustomClusterAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix+"-worker",
				terraformConfig.AWSConfig.AMI, terraformConfig.AWSConfig.AWSInstanceType, terratestConfig.WorkerCount)
		}
	case vsphereDefaults.Vsphere:
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

		vsphere.CreateVsphereVirtualMachine(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
	}

	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) {
		rootBody.AppendNewline()
		aws.CreateWindowsAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
	}

	rootBody.AppendNewline()

	SetRancher2ClusterV2(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	nullresource.CustomNullResource(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	return newFile, file, nil
}
