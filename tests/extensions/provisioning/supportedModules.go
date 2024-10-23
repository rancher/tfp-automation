package provisioning

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
)

// SupportedModules is a function that will check if the user-inputted module is supported.
func SupportedModules(terraformOptions *terraform.Options) bool {
	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(configs.Terraform, terraformConfig)

	module := terraformConfig.Module
	supportedModules := []string{
		clustertypes.AKS,
		clustertypes.EKS,
		clustertypes.GKE,
		modules.AzureRKE1,
		modules.AzureRKE2,
		modules.AzureK3s,
		modules.EC2RKE1,
		modules.EC2RKE2,
		modules.EC2K3s,
		modules.LinodeRKE1,
		modules.LinodeRKE2,
		modules.LinodeK3s,
		modules.VsphereRKE1,
		modules.VsphereRKE2,
		modules.VsphereK3s,
		modules.CustomEC2RKE1,
		modules.CustomEC2RKE2,
		modules.CustomEC2K3s,
	}

	for _, supportedModule := range supportedModules {
		if module == supportedModule {
			return true
		}
	}

	return false
}
