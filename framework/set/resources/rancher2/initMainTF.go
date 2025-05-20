package rancher2

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
)

// InitializeMainTF is a function that will create a new main.tf file for the downstream Rancher cluster tests
func InitializeMainTF(terratestConfig *config.TerratestConfig) (*hclwrite.File, *hclwrite.Body, *os.File) {
	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	_, keyPath := SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		return nil, nil, nil
	}

	return newFile, rootBody, file
}
