package provisioning

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/tfp-automation/config"
)

// SupportedModules is a function that will check if the user-inputted module is supported.
func SupportedModules(terraformOptions *terraform.Options) bool {
	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(terraformFrameworkConfig, terraformConfig)

	module := terraformConfig.Module
	supportedModules := []string{aks, eks, gke, ec2RKE1, linodeRKE1, ec2RKE2, linodeRKE2, ec2K3s, linodeK3s}

	for _, supportedModule := range supportedModules {
		if module == supportedModule {
			return true
		}
	}

	return false
}
