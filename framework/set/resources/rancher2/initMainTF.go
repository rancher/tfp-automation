package rancher2

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
)

// InitializeMainTF is a function that will create a new main.tf file for the downstream Rancher cluster tests
func InitializeMainTF() (*hclwrite.File, *hclwrite.Body, *os.File) {
	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	keyPath := SetKeyPath(keypath.RancherKeyPath, "")
	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		return nil, nil, nil
	}

	return newFile, rootBody, file
}
