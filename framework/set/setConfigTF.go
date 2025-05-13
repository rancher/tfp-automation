package set

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	configuration "github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	custom "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke1"
	customV2 "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/provisioning/hosted"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported"
	nodedriver "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke1"
	nodedriverV2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/rbac"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
)

// ConfigTF is a function that will set the main.tf file based on the module type.
func ConfigTF(client *rancher.Client, testUser, testPassword string, rbacRole configuration.Role, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, isWindows, persistClusters, customModule bool,
	customClusterNames []string) ([]string, []string, error) {
	var err error

	if !persistClusters {
		newFile.Body().Clear()
	}

	if !strings.Contains(string(newFile.Bytes()), defaults.RequiredProviders) {
		newFile, rootBody = resources.SetProvidersAndUsersTF(testUser, testPassword, false, newFile, rootBody, configMap, customModule)
	}

	rootBody.AppendNewline()

	clusterNames := []string{}
	containsCustomModule := false

	for i, cattleConfig := range configMap {
		_, terraform, terratest := config.LoadTFPConfigs(cattleConfig)

		kubernetesVersion := terratest.KubernetesVersion
		nodePools := terratest.Nodepools
		psact := terratest.PSACT
		snapshotInput := terratest.SnapshotInput

		module := terraform.Module

		if strings.Contains(module, clustertypes.CUSTOM) {
			containsCustomModule = true
		}

		clusterNames = append(clusterNames, terraform.ResourcePrefix)

		if module == modules.CustomEC2RKE2 || module == modules.CustomEC2K3s || module == modules.CustomEC2RKE2Windows ||
			module == modules.AirgapRKE2 || module == modules.AirgapK3S || module == modules.AirgapRKE2Windows {
			customClusterNames = append(customClusterNames, terraform.ResourcePrefix)
		}

		switch {
		case module == clustertypes.AKS:
			newFile, file, err = hosted.SetAKS(terraform, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		case module == clustertypes.EKS:
			newFile, file, err = hosted.SetEKS(terraform, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		case module == clustertypes.GKE:
			newFile, file, err = hosted.SetGKE(terraform, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		case strings.Contains(module, clustertypes.RKE1) && !strings.Contains(module, defaults.Custom) && !strings.Contains(module, defaults.Import) && !strings.Contains(module, defaults.Airgap):
			newFile, file, err = nodedriver.SetRKE1(terraform, kubernetesVersion, psact, nodePools, snapshotInput, newFile, rootBody, file, rbacRole)
			if err != nil {
				return clusterNames, nil, err
			}

			if rbacRole != "" {
				newFile, rootBody, err = rbac.RoleCheck(client, newFile, rootBody, file, terraform, rbacRole, true)
				if err != nil {
					return clusterNames, nil, err
				}
			}
		case (strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S)) && !strings.Contains(module, defaults.Custom) && !strings.Contains(module, defaults.Import) && !strings.Contains(module, defaults.Airgap):
			newFile, file, err = nodedriverV2.SetRKE2K3s(client, terraform, kubernetesVersion, psact, nodePools, snapshotInput, newFile, rootBody, file, rbacRole)
			if err != nil {
				return clusterNames, nil, err
			}

			if rbacRole != "" {
				newFile, rootBody, err = rbac.RoleCheck(client, newFile, rootBody, file, terraform, rbacRole, false)
				if err != nil {
					return clusterNames, nil, err
				}
			}
		case strings.Contains(module, clustertypes.RKE1) && strings.Contains(module, defaults.Custom):
			newFile, file, err = custom.SetCustomRKE1(terraform, terratest, configMap, newFile, rootBody, file)
			if err != nil {
				return clusterNames, customClusterNames, err
			}
		case (strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S)) && strings.Contains(module, defaults.Custom):
			if !isWindows {
				newFile, file, err = customV2.SetCustomRKE2K3s(terraform, terratest, configMap, newFile, rootBody, file)
				if err != nil {
					return clusterNames, customClusterNames, err
				}
			}

			if isWindows {
				newFile, file, err = customV2.SetCustomRKE2Windows(terraform, terratest, configMap, newFile, rootBody, file)
				if err != nil {
					return clusterNames, customClusterNames, err
				}
			}
		case module == modules.AirgapRKE1:
			newFile, file, err = airgap.SetAirgapRKE1(terraform, terratest, configMap, newFile, rootBody, file)
			if err != nil {
				return clusterNames, customClusterNames, err
			}
		case module == modules.AirgapRKE2 || module == modules.AirgapK3S || module == modules.AirgapRKE2Windows:
			if !isWindows {
				newFile, file, err = airgap.SetAirgapRKE2K3s(terraform, terratest, configMap, newFile, rootBody, file)
				if err != nil {
					return clusterNames, customClusterNames, err
				}
			}

			if isWindows {
				newFile, file, err = airgap.SetAirgapRKE2Windows(terraform, terratest, configMap, newFile, rootBody, file)
				if err != nil {
					return clusterNames, customClusterNames, err
				}
			}
		case strings.Contains(module, clustertypes.RKE1) && strings.Contains(module, defaults.Import):
			_, _, err = imported.SetImportedRKE1(terraform, terratest, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		case (strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S)) && strings.Contains(module, defaults.Import):
			_, _, err = imported.SetImportedRKE2K3s(terraform, terratest, newFile, rootBody, file)
			if err != nil {
				return clusterNames, nil, err
			}
		default:
			logrus.Errorf("Unsupported module: %v", module)
		}

		if i == len(configMap)-1 && containsCustomModule && !strings.Contains(module, defaults.Airgap) && !isWindows {
			file, err = locals.SetLocals(rootBody, terraform, terratest, configMap, newFile, file, customClusterNames)
			rootBody.AppendNewline()
		}

		if strings.Contains(module, defaults.Airgap) || isWindows {
			localsBlock := newFile.Body().FirstMatchingBlock(defaults.Locals, nil)
			if localsBlock != nil {
				newFile.Body().RemoveBlock(localsBlock)
			}

			file, err = locals.SetLocals(rootBody, terraform, terratest, configMap, newFile, file, customClusterNames)
			rootBody.AppendNewline()
		}
	}

	// // This is needed to ensure there is no duplications in the main.tf file.
	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	file, err = os.Create(keyPath + configs.MainTF)
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
