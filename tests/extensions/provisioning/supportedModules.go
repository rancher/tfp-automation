package provisioning

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
)

// SupportedModules is a function that will check if the user-inputted module is supported.
func SupportedModules(terraformConfig *config.TerraformConfig, terraformOptions *terraform.Options, configMap []map[string]any) bool {
	var isSupported bool
	for _, cattleConfig := range configMap {
		tfConfig := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, tfConfig)

		module := tfConfig.Module

		isSupported = verifyModule(module)
	}

	return isSupported
}

func verifyModule(module string) bool {
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
		modules.HarvesterRKE1,
		modules.HarvesterRKE2,
		modules.HarvesterK3s,
		modules.LinodeRKE1,
		modules.LinodeRKE2,
		modules.LinodeK3s,
		modules.VsphereRKE1,
		modules.VsphereRKE2,
		modules.VsphereK3s,
		modules.CustomEC2RKE1,
		modules.CustomEC2RKE2,
		modules.CustomEC2K3s,
		modules.AirgapRKE1,
		modules.AirgapRKE2,
		modules.AirgapK3S,
		modules.ImportEC2RKE1,
		modules.ImportEC2RKE2,
		modules.ImportEC2K3s,
	}

	for _, supportedModule := range supportedModules {
		if module == supportedModule {
			return true
		}
	}

	return false
}
