package cli

import (
	"os"
	"testing"

	"github.com/rancher/tfp-automation/tests/infrastructure/clusters"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
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

var setupRancherFuncs = map[string]func(*testing.T, string) error{
	"--airgap:fresh":    ranchers.CreateAirgapRancher,
	"--airgap:upgrade":  ranchers.UpgradingAirgapRancher,
	"--dualstack:fresh": ranchers.CreateDualStackRancher,
	"--ipv6:fresh":      ranchers.CreateIPv6Rancher,
	"--normal:fresh":    ranchers.CreateRancher,
	"--normal:upgrade":  ranchers.UpgradingRancher,
	"--proxy:fresh":     ranchers.CreateProxyRancher,
	"--proxy:upgrade":   ranchers.UpgradingProxyRancher,
	"--registry:fresh":  ranchers.CreateRegistryRancher,
}

// RunCLI is a function that runs the CLI setup based on command-line arguments
func RunCLI() int {
	t := &testing.T{}
	key := os.Args[1]

	// If there are two arguments, we assume this is a Rancher setup. If only one, it's a cluster setup.
	if len(os.Args) > 2 {
		key += ":" + os.Args[2]

		if setupFunc, ok := setupRancherFuncs[key]; ok {
			err := setupFunc(t, "")
			require.NoError(t, err)

			return 0
		}
	} else {
		if setupFunc, ok := setupClusterFuncs[key]; ok {
			err := setupFunc(t, "")
			require.NoError(t, err)

			return 0
		}
	}

	return 0
}
