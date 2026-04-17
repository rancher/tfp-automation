package set

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported/rke2k3s"
)

// ImportedClusters is a function that will set the imported clusters in the main.tf file.
func ImportedClusters(client *rancher.Client, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File,
	isWindows bool) (*hclwrite.File, *os.File, error) {
	var err error

	newFile, file, err = rke2k3s.SetImportedRKE2K3s(terraformConfig, terratestConfig, newFile, rootBody, file)
	if err != nil {
		return newFile, file, err
	}

	return newFile, file, nil
}
