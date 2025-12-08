package web

import (
	"os"
	"strings"
	"testing"
	"time"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/tests/infrastructure/clusters"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	share "github.com/rancher/tfp-automation/tests/infrastructure/state"
)

var setupClusterFuncs = map[string]func(*testing.T, string) error{
	"airgap-rke2": clusters.CreateAirgappedRKE2Cluster,
	"dual-rke2":   clusters.CreateDualStackRKE2Cluster,
	"dual-k3s":    clusters.CreateDualStackK3SCluster,
	"ipv6-rke2":   clusters.CreateIPv6RKE2Cluster,
	"ipv6-k3s":    clusters.CreateIPv6K3SCluster,
	"normal-rke2": clusters.CreateRKE2Cluster,
	"normal-k3s":  clusters.CreateK3SCluster,
}

var setupRancherFuncs = map[string]map[string]func(*testing.T, string) error{
	"airgap": {
		"fresh":   ranchers.CreateAirgapRancher,
		"upgrade": ranchers.UpgradingAirgapRancher,
	},
	"dual": {
		"fresh": ranchers.CreateDualStackRancher,
	},
	"ipv6": {
		"fresh": ranchers.CreateIPv6Rancher,
	},
	"normal": {
		"fresh":   ranchers.CreateRancher,
		"upgrade": ranchers.UpgradingRancher,
	},
	"proxy": {
		"fresh":   ranchers.CreateProxyRancher,
		"upgrade": ranchers.UpgradingProxyRancher,
	},
	"registry": {
		"fresh": ranchers.CreateRegistryRancher,
	},
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
