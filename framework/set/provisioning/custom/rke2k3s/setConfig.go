package rke2k3s

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/vsphere"
	"github.com/sirupsen/logrus"
)

// SetCustomRKE2K3s is a function that will set the custom RKE2/K3s cluster configurations in the main.tf file.
func SetCustomRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	if strings.Contains(terraformConfig.Module, defaults.Custom) && !strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
	} else if strings.Contains(terraformConfig.Module, modules.CustomVsphereRKE2) || strings.Contains(terraformConfig.Module, modules.CustomVsphereK3s) {
		dataCenterExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereDatacenter + `.` + defaults.VsphereDatacenter + `.id`)
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

	if strings.Contains(terraformConfig.Module, modules.CustomEC2RKE2Windows) {
		rootBody.AppendNewline()
		aws.CreateWindowsAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
	}

	rootBody.AppendNewline()

	SetRancher2ClusterV2(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	nullresource.CustomNullResource(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom RKE2/K3s configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}
