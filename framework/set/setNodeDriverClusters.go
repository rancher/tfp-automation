package set

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	configuration "github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke1"
	"github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/rbac"
)

// NodeDriverClusters is a function that will set the node driver clusters in the main.tf file.
func NodeDriverClusters(client *rancher.Client, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	rbacRole configuration.Role, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	var err error

	if strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		newFile, file, err = rke1.SetRKE1(terraformConfig, terratestConfig, newFile, rootBody, file, rbacRole)
		if err != nil {
			return newFile, file, err
		}

		if rbacRole != "" {
			newFile, rootBody, err = rbac.RoleCheck(client, newFile, rootBody, file, terraformConfig, rbacRole, true)
			if err != nil {
				return newFile, file, err
			}
		}
	}

	if strings.Contains(terraformConfig.Module, clustertypes.RKE2) || strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		newFile, file, err = rke2k3s.SetRKE2K3s(terraformConfig, terratestConfig, newFile, rootBody, file, rbacRole)
		if err != nil {
			return newFile, file, err
		}

		if rbacRole != "" {
			newFile, rootBody, err = rbac.RoleCheck(client, newFile, rootBody, file, terraformConfig, rbacRole, false)
			if err != nil {
				return newFile, file, err
			}
		}
	}

	return newFile, file, nil
}
