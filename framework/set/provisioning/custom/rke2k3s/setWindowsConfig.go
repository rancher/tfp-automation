package rke2k3s

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/nullresource"
	"github.com/sirupsen/logrus"
)

// SetCustomRKE2Windows is a function that will set the custom RKE2 cluster configurations in the main.tf file.
func SetCustomRKE2Windows(client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	nullresource.SetWindowsNullResource(rootBody, terraformConfig)
	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write custom Windows RKE2 configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
