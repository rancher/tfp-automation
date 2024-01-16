package provisioning

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
)

// SupportedModules is a function that will check if the user-inputted module is supported.
func SupportedModules(terraformOptions *terraform.Options) bool {
	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(terraformFrameworkConfig, terraformConfig)

	module := terraformConfig.Module
	supportedModules := []string{aks, eks, gke, azureRKE1, azureRKE2, azureK3s, ec2RKE1, linodeRKE1, ec2RKE2, linodeRKE2, ec2K3s, linodeK3s, vsphereRKE1, vsphereRKE2, vsphereK3s}

	for _, supportedModule := range supportedModules {
		if module == supportedModule {
			return true
		}
	}

	return false
}
