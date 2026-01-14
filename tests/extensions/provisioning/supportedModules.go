package provisioning

import (
	"slices"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
)

// DownstreamClusterModules is a function that will return the correct module names based on the provider.
func DownstreamClusterModules(terraformConfig *config.TerraformConfig) (string, string, string, string) {
	var rke2Module, rke2Windows2019, rke2Windows2022, k3sModule string
	switch terraformConfig.DownstreamClusterProvider {
	case defaults.Aws:
		rke2Module = modules.EC2RKE2
		rke2Windows2019 = modules.CustomEC2RKE2Windows2019
		rke2Windows2022 = modules.CustomEC2RKE2Windows2022
		k3sModule = modules.EC2K3s
	case defaults.Azure:
		rke2Module = modules.AzureRKE2
		k3sModule = modules.AzureK3s
	case defaults.Linode:
		rke2Module = modules.LinodeRKE2
		k3sModule = modules.LinodeK3s
	case defaults.Vsphere:
		rke2Module = modules.VsphereRKE2
		k3sModule = modules.VsphereK3s
	default:
		panic("Unsupported provider: " + terraformConfig.DownstreamClusterProvider)
	}

	return rke2Module, rke2Windows2019, rke2Windows2022, k3sModule
}

// ImportedClusterModules is a function that will return the correct module names based on the provider.
func ImportedClusterModules(terraformConfig *config.TerraformConfig) (string, string, string, string) {
	var rke2Module, rke2Windows2019, rke2Windows2022, k3sModule string
	switch terraformConfig.DownstreamClusterProvider {
	case defaults.Aws:
		rke2Module = modules.ImportEC2RKE2
		rke2Windows2019 = modules.ImportEC2RKE2Windows2019
		rke2Windows2022 = modules.ImportEC2RKE2Windows2022
		k3sModule = modules.ImportEC2K3s
	case defaults.Vsphere:
		rke2Module = modules.ImportVsphereRKE2
		k3sModule = modules.ImportVsphereK3s
	default:
		panic("Unsupported provider: " + terraformConfig.DownstreamClusterProvider)
	}

	return rke2Module, rke2Windows2019, rke2Windows2022, k3sModule
}

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
		modules.AKS,
		modules.EKS,
		modules.GKE,
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
		modules.CustomEC2RKE2Windows2019,
		modules.CustomEC2RKE2Windows2022,
		modules.CustomEC2K3s,
		modules.CustomVsphereRKE1,
		modules.CustomVsphereRKE2,
		modules.CustomVsphereK3s,
		modules.AirgapRKE1,
		modules.AirgapRKE2,
		modules.AirgapRKE2Windows2019,
		modules.AirgapRKE2Windows2022,
		modules.AirgapK3S,
		modules.ImportEC2RKE1,
		modules.ImportEC2RKE2,
		modules.ImportEC2RKE2Windows2019,
		modules.ImportEC2RKE2Windows2022,
		modules.ImportEC2K3s,
		modules.ImportVsphereRKE1,
		modules.ImportVsphereRKE2,
		modules.ImportVsphereK3s,
	}

	return slices.Contains(supportedModules, module)
}
