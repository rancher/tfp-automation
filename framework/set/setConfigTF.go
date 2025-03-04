package set

import (
	"os"
	"strings"

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
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
)

// ConfigTF is a function that will set the main.tf file based on the module type.
func ConfigTF(client *rancher.Client, testUser, testPassword string, rbacRole configuration.Role, configMap []map[string]any) ([]string, error) {
	var file *os.File
	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)

	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return nil, err
	}

	defer file.Close()
	newFile, rootBody := resources.SetProvidersAndUsersTF(testUser, testPassword, false, configMap)

	rootBody.AppendNewline()

	clusterNames := []string{}
	customClusterNames := []string{}
	containsCustomModule := false

	for i, cattleConfig := range configMap {
		rancherConfig, terraform, terratest := config.LoadTFPConfigs(cattleConfig)

		kubernetesVersion := terratest.KubernetesVersion
		nodePools := terratest.Nodepools
		psact := terratest.PSACT
		snapshotInput := terratest.SnapshotInput

		module := terraform.Module

		if strings.Contains(module, clustertypes.CUSTOM) {
			containsCustomModule = true
		}

		clusterNames = append(clusterNames, terraform.ResourcePrefix)

		if module == modules.CustomEC2RKE2 || module == modules.CustomEC2K3s {
			customClusterNames = append(customClusterNames, terraform.ResourcePrefix)
		}

		switch {
		case module == clustertypes.AKS:
			file, err = hosted.SetAKS(terraform, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == clustertypes.EKS:
			file, err = hosted.SetEKS(terraform, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == clustertypes.GKE:
			file, err = hosted.SetGKE(terraform, kubernetesVersion, nodePools, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case strings.Contains(module, clustertypes.RKE1) && !strings.Contains(module, defaults.Custom) && !strings.Contains(module, defaults.Import) && !strings.Contains(module, defaults.Airgap):
			file, err = nodedriver.SetRKE1(terraform, kubernetesVersion, psact, nodePools, snapshotInput, newFile, rootBody, file, rbacRole)
			if err != nil {
				return clusterNames, err
			}
		case (strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S)) && !strings.Contains(module, defaults.Custom) && !strings.Contains(module, defaults.Import) && !strings.Contains(module, defaults.Airgap):
			file, err = nodedriverV2.SetRKE2K3s(client, terraform, kubernetesVersion, psact, nodePools, snapshotInput, newFile, rootBody, file, rbacRole)
			if err != nil {
				return clusterNames, err
			}
		case module == modules.CustomEC2RKE1:
			file, err = custom.SetCustomRKE1(rancherConfig, terraform, terratest, configMap, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == modules.CustomEC2RKE2 || module == modules.CustomEC2K3s:
			file, err = customV2.SetCustomRKE2K3s(rancherConfig, terraform, terratest, configMap, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == modules.AirgapRKE1:
			_, err = airgap.SetAirgapRKE1(rancherConfig, terraform, terratest, configMap, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == modules.AirgapRKE2 || module == modules.AirgapK3S:
			_, err = airgap.SetAirgapRKE2K3s(rancherConfig, terraform, terratest, configMap, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == modules.ImportEC2RKE1:
			_, err = imported.SetImportedRKE1(rancherConfig, terraform, terratest, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		case module == modules.ImportEC2RKE2 || module == modules.ImportEC2K3s:
			_, err = imported.SetImportedRKE2K3s(rancherConfig, terraform, terratest, newFile, rootBody, file)
			if err != nil {
				return clusterNames, err
			}
		default:
			logrus.Errorf("Unsupported module: %v", module)
		}

		if i == len(configMap)-1 && containsCustomModule {
			file, err = locals.SetLocals(rootBody, terraform, configMap, newFile, file, customClusterNames)
		}
	}

	keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath)

	file, err = os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return clusterNames, err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return clusterNames, err
	}

	return clusterNames, nil
}
