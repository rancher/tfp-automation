package web

import (
	"os"
	"strings"
	"testing"
	"time"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/tests/infrastructure/clusters"

	setupairgap "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/airgap"
	setupdualstack "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/dualstack"
	setupipv6 "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/ipv6"
	setupproxy "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/proxy"
	setupregistry "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/registry"
	setupstandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/standard"
	upgradeairgap "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/airgap"
	upgradeproxy "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/proxy"
	upgradestandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/standard"
	"github.com/rancher/tfp-automation/tests/infrastructure/registries"
	share "github.com/rancher/tfp-automation/tests/infrastructure/state"
)

var setupClusterFuncs = map[string]func(*testing.T, string) error{
	"airgap-rke2": clusters.CreateAirgappedRKE2Cluster,
	"airgap-k3s":  clusters.CreateAirgappedK3SCluster,
	"dual-rke2":   clusters.CreateDualStackRKE2Cluster,
	"dual-k3s":    clusters.CreateDualStackK3SCluster,
	"ipv6-rke2":   clusters.CreateIPv6RKE2Cluster,
	"ipv6-k3s":    clusters.CreateIPv6K3SCluster,
	"normal-rke2": clusters.CreateRKE2Cluster,
	"normal-k3s":  clusters.CreateK3SCluster,
	"proxy-rke2":  clusters.CreateProxyRKE2Cluster,
	"proxy-k3s":   clusters.CreateProxyK3SCluster,
}

var setupRancherFuncs = map[string]map[string]func(*testing.T, string) error{
	"airgap": {
		"fresh": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return setupairgap.CreateAirgapRancher(t, provider, cattleConfig)
		},
		"upgrade": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return upgradeairgap.UpgradingAirgapRancher(t, provider, cattleConfig)
		},
	},
	"dual": {
		"fresh": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return setupdualstack.CreateDualStackRancher(t, provider, cattleConfig)
		},
	},
	"ipv6": {
		"fresh": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return setupipv6.CreateIPv6Rancher(t, provider, cattleConfig)
		},
	},
	"normal": {
		"fresh": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return setupstandard.CreateRancher(t, provider, cattleConfig)
		},
		"upgrade": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return upgradestandard.UpgradingRancher(t, provider, cattleConfig)
		},
	},
	"proxy": {
		"fresh": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return setupproxy.CreateProxyRancher(t, provider, cattleConfig)
		},
		"upgrade": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return upgradeproxy.UpgradingProxyRancher(t, provider, cattleConfig)
		},
	},
	"registry": {
		"fresh": func(t *testing.T, provider string) error {
			cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
			return setupregistry.CreateRegistryRancher(t, provider, cattleConfig)
		},
	},
}

var setupRegistryFuncs = map[string]func(*testing.T, string) error{
	"registries-all":    registries.SetupAllRegistries,
	"registries-auth":   registries.SetupAuthenticatedRegistry,
	"registries-unauth": registries.SetupUnauthenticatedRegistry,
	"registries-ecr":    registries.SetupECR,
}

// RunClusterSetupWeb is a function that runs the standalone cluster web application setup
func RunClusterSetupWeb(provider, providerversion, clustertype string) error {
	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	os.Setenv("CLOUD_PROVIDER_VERSION", providerversion)

	t := &testing.T{}

	share.State.Mutex.Lock()

	share.State.StageMsg = strings.Join(share.ClusterStageMessage, "\n")
	share.State.ErrorMsg = ""
	share.State.Mutex.Unlock()

	var setupErr error
	if setupFunc, ok := setupClusterFuncs[clustertype]; ok {
		setupErr = setupFunc(t, provider)
	}

	if setupErr != nil {
		share.State.Mutex.Lock()
		share.State.ErrorMsg = setupErr.Error()
		share.State.StageMsg = "Unable to create cluster. See error below:"
		share.State.Mutex.Unlock()

		go func() {
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}()

		return setupErr
	}

	share.State.StageMsg = "Cluster is ready! Please find the kubeconfig in the initial cluster node at /home/" +
		terraformConfig.Standalone.OSUser + "/.kube/config"

	go func() {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}()

	return nil
}

// RunRancherSetupWeb is a function that runs the Rancher web application setup
func RunRancherSetupWeb(provider, providerversion, ranchertype, installtype string) error {
	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	os.Setenv("CLOUD_PROVIDER_VERSION", providerversion)

	t := &testing.T{}

	share.State.Mutex.Lock()
	share.State.StageMsg = strings.Join(share.RancherStageMessage, "\n")
	share.State.ErrorMsg = ""
	share.State.Mutex.Unlock()

	var setupErr error
	if installMap, ok := setupRancherFuncs[ranchertype]; ok {
		if setupFunc, ok := installMap[installtype]; ok {
			setupErr = setupFunc(t, provider)
		}
	}

	if setupErr != nil {
		share.State.Mutex.Lock()
		share.State.ErrorMsg = setupErr.Error()
		share.State.StageMsg = "Unable to create Rancher. See error below:"
		share.State.Mutex.Unlock()

		go func() {
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}()

		return setupErr
	}

	url := "https://" + terraformConfig.Standalone.RancherHostname
	share.State.StageMsg = strings.TrimSpace(url)

	go func() {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}()

	return nil
}

// RunRegistrySetupWeb is a function that runs the standalone registry web application setup
func RunRegistrySetupWeb(provider, providerversion, registrytype string) error {
	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	os.Setenv("CLOUD_PROVIDER_VERSION", providerversion)

	t := &testing.T{}

	share.State.Mutex.Lock()

	share.State.StageMsg = strings.Join(share.RegistryStageMessage, "\n")
	share.State.ErrorMsg = ""
	share.State.Mutex.Unlock()

	var setupErr error
	if setupFunc, ok := setupRegistryFuncs[registrytype]; ok {
		setupErr = setupFunc(t, provider)
	}

	if setupErr != nil {
		share.State.Mutex.Lock()
		share.State.ErrorMsg = setupErr.Error()
		share.State.StageMsg = "Unable to create registry. See error below:"
		share.State.Mutex.Unlock()

		go func() {
			time.Sleep(30 * time.Second)
			os.Exit(1)
		}()

		return setupErr
	}

	share.State.StageMsg = "Registry is ready! Please go to AWS Management Console and find it with prefix " + terraformConfig.ResourcePrefix

	go func() {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}()

	return nil
}
