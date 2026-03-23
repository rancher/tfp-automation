package set

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	configuration "github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver"
	"github.com/rancher/tfp-automation/framework/set/rbac"
)

// NodeDriverClusters is a function that will set the node driver clusters in the main.tf file.
func NodeDriverClusters(client *rancher.Client, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	rbacRole configuration.Role, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	var err error

	newFile, file, err = nodedriver.SetRKE2K3s(terraformConfig, terratestConfig, newFile, rootBody, file, rbacRole)
	if err != nil {
		return newFile, file, err
	}

	if rbacRole != "" {
		newFile, rootBody, err = rbac.RoleCheck(client, newFile, rootBody, file, terraformConfig, rbacRole)
		if err != nil {
			return newFile, file, err
		}
	}

	return newFile, file, nil
}
