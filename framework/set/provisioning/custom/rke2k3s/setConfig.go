package rke2k3s

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
)

// SetCustomRKE2K3s is a function that will set the custom RKE2/K3s cluster configurations in the main.tf file.
func SetCustomRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)

	if strings.Contains(terraformConfig.Module, modules.CustomEC2RKE2Windows) {
		rootBody.AppendNewline()
		aws.CreateWindowsAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
	}

	rootBody.AppendNewline()

	SetRancher2ClusterV2(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	nullresource.SetNullResource(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom RKE2/K3s configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}
