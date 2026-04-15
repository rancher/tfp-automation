package set

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
)

// ConfigTF sets the main.tf file based on the module type.
func ConfigTF(client *rancher.Client, rancherConfig *rancher.Config, terratestConfig *config.TerratestConfig, testUser, testPassword string,
	rbacRole config.Role, configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows, persistClusters,
	customModule bool, customClusterNames []string, nestedRancherModuleDir string) ([]string, []string, error) {
	var err error

	clusterNames := []string{}
	containsCustomModule := false

	if !persistClusters {
		newFile.Body().Clear()
	}

	if !strings.Contains(string(newFile.Bytes()), general.RequiredProviders) {
		newFile, rootBody = rancher2.SetProvidersAndUsersTF(rancherConfig, testUser, testPassword, false, newFile, rootBody, configMap, customModule)
	}

	rootBody.AppendNewline()

	for i, cattleConfig := range configMap {
		_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

		if strings.Contains(terraformConfig.Module, clustertypes.CUSTOM) || strings.Contains(terraformConfig.Module, general.Airgap) {
			containsCustomModule = true
		}

		clusterNames = append(clusterNames, terraformConfig.ResourcePrefix)

		if strings.Contains(terraformConfig.Module, general.Custom) || strings.Contains(terraformConfig.Module, general.Airgap) {
			customClusterNames = append(customClusterNames, terraformConfig.ResourcePrefix)
		}

		if terraformConfig.Module == modules.HostedAzureAKS || terraformConfig.Module == modules.HostedAWSEKS || terraformConfig.Module == modules.HostedGoogleGKE {
			newFile, file, err = HostedClusters(terraformConfig, terratestConfig, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		}

		if !strings.Contains(terraformConfig.Module, general.Custom) && !strings.Contains(terraformConfig.Module, general.Import) && !strings.Contains(terraformConfig.Module, general.Airgap) &&
			terraformConfig.Module != modules.HostedAzureAKS && terraformConfig.Module != modules.HostedAWSEKS && terraformConfig.Module != modules.HostedGoogleGKE {
			newFile, file, err = NodeDriverClusters(client, terraformConfig, terratestConfig, rbacRole, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		}

		if strings.Contains(terraformConfig.Module, general.Custom) {
			newFile, file, err = CustomClusters(client, terraformConfig, terratestConfig, newFile, rootBody, file, configMap, isWindows)
			if err != nil {
				return clusterNames, nil, err
			}
		}

		if strings.Contains(terraformConfig.Module, general.Airgap) {
			newFile, file, err = AirgapClusters(client, terraformConfig, terratestConfig, newFile, rootBody, file, configMap, isWindows)
			if err != nil {
				return clusterNames, nil, err
			}
		}

		if strings.Contains(terraformConfig.Module, general.Import) {
			newFile, file, err = ImportedClusters(client, terraformConfig, terratestConfig, newFile, rootBody, file, configMap, isWindows)
			if err != nil {
				return clusterNames, nil, err
			}
		}

		if i == len(configMap)-1 && containsCustomModule {
			localsBlock := newFile.Body().FirstMatchingBlock(general.Locals, nil)
			if localsBlock != nil {
				newFile.Body().RemoveBlock(localsBlock)
			}

			file, err = locals.SetLocals(rootBody, terraformConfig, terratestConfig, configMap, newFile, file, customClusterNames)
			rootBody.AppendNewline()
		}
	}

	// This is needed to ensure there is no duplications in the main.tf file.
	mainTFPath := nestedRancherModuleDir + configs.MainTF
	file, err = os.Create(mainTFPath)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return clusterNames, customClusterNames, err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write configurations to main.tf file. Error: %v", err)
		return clusterNames, customClusterNames, err
	}

	return clusterNames, customClusterNames, nil
}
