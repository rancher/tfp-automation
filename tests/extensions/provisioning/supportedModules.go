package provisioning

import (
	"slices"
	"strings"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/harvester"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
)

// DownstreamClusterModules is a function that will return the correct module names based on the provider.
func DownstreamClusterModules(terraformConfig *config.TerraformConfig) (string, string, string, string) {
	var rke2Module, rke2Windows2019, rke2Windows2022, k3sModule string
	switch terraformConfig.DownstreamClusterProvider {
	case aws.Aws:
		rke2Module = modules.NodeDriverAWSRKE2
		rke2Windows2019 = modules.CustomAWSRKE2Windows2019
		rke2Windows2022 = modules.CustomAWSRKE2Windows2022
		k3sModule = modules.NodeDriverAWSK3S
	case azure.Azure:
		rke2Module = modules.NodeDriverAzureRKE2
		k3sModule = modules.NodeDriverAzureK3S
	case google.Google:
		rke2Module = modules.NodeDriverGoogleRKE2
		k3sModule = modules.NodeDriverGoogleK3S
	case harvester.Harvester:
		rke2Module = modules.NodeDriverHarvesterRKE2
		k3sModule = modules.NodeDriverHarvesterK3S
	case linode.Linode:
		rke2Module = modules.NodeDriverLinodeRKE2
		k3sModule = modules.NodeDriverLinodeK3S
	case vsphere.Vsphere:
		rke2Module = modules.NodeDriverVsphereRKE2
		k3sModule = modules.NodeDriverVsphereK3S
	default:
		panic("Unsupported provider: " + terraformConfig.DownstreamClusterProvider)
	}

	return rke2Module, rke2Windows2019, rke2Windows2022, k3sModule
}

// ImportedClusterModules is a function that will return the correct module names based on the provider.
func ImportedClusterModules(terraformConfig *config.TerraformConfig) (string, string, string, string) {
	var rke2Module, rke2Windows2019, rke2Windows2022, k3sModule string
	switch terraformConfig.DownstreamClusterProvider {
	case aws.Aws:
		rke2Module = modules.ImportedAWSRKE2
		rke2Windows2019 = modules.ImportedAWSRKE2Windows2019
		rke2Windows2022 = modules.ImportedAWSRKE2Windows2022
		k3sModule = modules.ImportedAWSK3S
	case vsphere.Vsphere:
		rke2Module = modules.ImportedVsphereRKE2
		k3sModule = modules.ImportedVsphereK3S
	default:
		panic("Unsupported provider: " + terraformConfig.DownstreamClusterProvider)
	}

	return rke2Module, rke2Windows2019, rke2Windows2022, k3sModule
}

func IsImportedModule(module string) bool {
	if strings.Contains(module, "import") {
		return true
	}

	return false
}

func IsHostedModule(module string) bool {
	return module == modules.HostedAzureAKS || module == modules.HostedAWSEKS || module == modules.HostedGoogleGKE
}

// SupportedModules is a function that will check if the user-inputted module is supported.
func SupportedModules(terraformConfig *config.TerraformConfig) bool {
	if terraformConfig == nil {
		return false
	}

	return verifyModule(terraformConfig.Module)
}

func verifyModule(module string) bool {
	supportedModules := []string{
		modules.HostedAzureAKS,
		modules.HostedAWSEKS,
		modules.HostedGoogleGKE,
		modules.NodeDriverAzureRKE2,
		modules.NodeDriverAzureK3S,
		modules.NodeDriverAWSRKE2,
		modules.NodeDriverAWSK3S,
		modules.NodeDriverGoogleRKE2,
		modules.NodeDriverGoogleK3S,
		modules.NodeDriverHarvesterRKE2,
		modules.NodeDriverHarvesterK3S,
		modules.NodeDriverLinodeRKE2,
		modules.NodeDriverLinodeK3S,
		modules.NodeDriverVsphereRKE2,
		modules.NodeDriverVsphereK3S,
		modules.CustomAWSRKE2,
		modules.CustomAWSRKE2Windows2019,
		modules.CustomAWSRKE2Windows2022,
		modules.CustomAWSK3S,
		modules.CustomVsphereRKE2,
		modules.CustomVsphereK3S,
		modules.AirgapAWSRKE2,
		modules.AirgapAWSRKE2Windows2022,
		modules.AirgapAWSK3S,
		modules.ImportedAWSRKE2,
		modules.ImportedAWSRKE2Windows2019,
		modules.ImportedAWSRKE2Windows2022,
		modules.ImportedAWSK3S,
		modules.ImportedVsphereRKE2,
		modules.ImportedVsphereK3S,
	}

	return slices.Contains(supportedModules, module)
}
