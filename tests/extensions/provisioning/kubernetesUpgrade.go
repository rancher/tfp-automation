package provisioning

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/tfp-automation/config"
	set "github.com/rancher/tfp-automation/framework/set/provisioning"
	"github.com/stretchr/testify/require"
)

const (
	RKE1 = "-rancher"
	RKE2 = "rke2"
	K3S  = "k3s"
)

// KubernetesUpgrade is a function that will run terraform apply and uprade the Kubernetes version of the provisioned cluster.
func KubernetesUpgrade(t *testing.T, client *rancher.Client, clusterName string, terraformOptions *terraform.Options, clusterConfig *config.TerratestConfig) {
	if clusterConfig.UpgradedKubernetesVersion == "" {
		if strings.Contains(clusterConfig.KubernetesVersion, RKE1) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE1ClusterType.String(), nil)
			clusterConfig.KubernetesVersion = defaultVersion[0]
			require.NoError(t, err)
		} else if strings.Contains(clusterConfig.KubernetesVersion, RKE2) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
			clusterConfig.KubernetesVersion = defaultVersion[0]
			require.NoError(t, err)
		} else if strings.Contains(clusterConfig.KubernetesVersion, K3S) {
			defaultVersion, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
			clusterConfig.KubernetesVersion = defaultVersion[0]
			require.NoError(t, err)
		}
	} else {
		clusterConfig.KubernetesVersion = clusterConfig.UpgradedKubernetesVersion
	}

	err := set.SetConfigTF(clusterConfig, clusterName)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)
}
