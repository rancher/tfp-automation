package provisioning

import (
	"slices"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
)

// SupportedModules is a function that will check if the user-inputted module is supported.
func SupportedModules(terraformOptions *terraform.Options, configMap []map[string]any) bool {
	var isSupported bool
	for _, cattleConfig := range configMap {
		tfConfig := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, tfConfig)

		isSupported = verifyModule(tfConfig.Module)
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
		modules.CustomEC2RKE2Windows,
		modules.CustomEC2K3s,
		modules.AirgapRKE1,
		modules.AirgapRKE2,
		modules.AirgapK3S,
		modules.ImportEC2RKE1,
		modules.ImportEC2RKE2,
		modules.ImportEC2K3s,
	}

	return slices.Contains(supportedModules, module)
}
