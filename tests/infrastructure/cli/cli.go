package cli

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/tests/infrastructure/clusters"

	setupairgap "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/airgap"
	setupdualstack "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/dualstack"
	setuphosted "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/hosted"
	setupipv6 "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/ipv6"
	setupproxy "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/proxy"
	setupregistry "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/registry"
	setupstandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/standard"
	upgradeairgap "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/airgap"
	upgradedualstack "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/dualstack"
	upgradeipv6 "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/ipv6"
	upgradeproxy "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/proxy"
	upgradestandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/standard"
	"github.com/rancher/tfp-automation/tests/infrastructure/registries"
	"github.com/stretchr/testify/require"
)

var setupClusterFuncs = map[string]func(*testing.T, string) error{
	"--airgap-rke2": clusters.CreateAirgappedRKE2Cluster,
	"--dual-rke2":   clusters.CreateDualStackRKE2Cluster,
	"--dual-k3s":    clusters.CreateDualStackK3SCluster,
	"--ipv6-rke2":   clusters.CreateIPv6RKE2Cluster,
	"--ipv6-k3s":    clusters.CreateIPv6K3SCluster,
	"--normal-rke2": clusters.CreateRKE2Cluster,
	"--normal-k3s":  clusters.CreateK3SCluster,
}

var setupRancherFuncs = map[string]func(*testing.T, string, map[string]any) error{
	"--airgap:fresh":   setupairgap.CreateAirgapRancher,
	"--airgap:upgrade": upgradeairgap.UpgradingAirgapRancher,
	"--dual:fresh":     setupdualstack.CreateDualStackRancher,
	"--dual:upgrade":   upgradedualstack.UpgradingDualStackRancher,
	"--hosted:fresh":   setuphosted.CreateHostedClusterRancher,
	"--ipv6:fresh":     setupipv6.CreateIPv6Rancher,
	"--ipv6:upgrade":   upgradeipv6.UpgradingIPv6Rancher,
	"--normal:fresh":   setupstandard.CreateRancher,
	"--normal:upgrade": upgradestandard.UpgradingRancher,
	"--proxy:fresh":    setupproxy.CreateProxyRancher,
	"--proxy:upgrade":  upgradeproxy.UpgradingProxyRancher,
	"--registry:fresh": setupregistry.CreateRegistryRancher,
}

var setupRegistryFuncs = map[string]func(*testing.T, string) error{
	"--registries-all":     registries.SetupAllRegistries,
	"--registries-auth":    registries.SetupAuthenticatedRegistry,
	"--registries-nonauth": registries.SetupNonAuthenticatedRegistry,
	"--registries-ecr":     registries.SetupECR,
}

// RunCLI is a function that runs the CLI setup based on command-line arguments
func RunCLI() int {
	t := &testing.T{}
	key := os.Args[1]

	// If there are two arguments, we assume this is a Rancher setup. If only one, it's a cluster or registry setup.
	if len(os.Args) > 2 {
		key += ":" + os.Args[2]

		if setupFunc, ok := setupRancherFuncs[key]; ok {
			cattleConfig := config.LoadConfigFromFile(os.Getenv(config.ConfigEnvironmentKey))
			err := setupFunc(t, "", cattleConfig)
			require.NoError(t, err)

			return 0
		}
	} else {
		if setupFunc, ok := setupClusterFuncs[key]; ok {
			err := setupFunc(t, "")
			require.NoError(t, err)

			return 0
		}

		if setupFunc, ok := setupRegistryFuncs[key]; ok {
			err := setupFunc(t, "")
			require.NoError(t, err)

			return 0
		}
	}

	return 0
}
