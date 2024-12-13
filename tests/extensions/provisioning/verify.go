package provisioning

import (
	"strings"
	"testing"

	clusterActions "github.com/rancher/rancher/tests/v2/actions/clusters"
	"github.com/rancher/rancher/tests/v2/actions/psact"
	"github.com/rancher/shepherd/clients/rancher"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	waitState "github.com/rancher/tfp-automation/framework/wait/state"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// VerifyCluster validates that a downstream cluster and its resources are in a good state, matching a given config.
func VerifyCluster(t *testing.T, client *rancher.Client, clusterName string, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) {
	var expectedKubernetesVersion string
	module := terraformConfig.Module
	expectedKubernetesVersion = checkExpectedKubernetesVersion(t, terratestConfig, expectedKubernetesVersion, module)

	clusterID, err := clusterExtensions.GetClusterIDByName(client, clusterName)
	require.NoError(t, err)

	logrus.Infof("Waiting for cluster %v to be in an active state...", clusterName)
	if err := waitState.IsActiveCluster(client, clusterID); err != nil {
		require.NoError(t, err)
	}

	if err := waitState.AreNodesActive(client, clusterID); err != nil {
		require.NoError(t, err)
	}

	cluster, err := client.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	// EKS is formatted this way due to EKS formatting Kubernetes versions with a random string of letters after the version.
	if module == clustertypes.EKS {
		assert.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion[1:5])
	} else {
		assert.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)
	}

	clusterToken, err := clusterActions.CheckServiceAccountTokenSecret(client, cluster.Name)
	require.NoError(t, err)
	assert.NotEmpty(t, clusterToken)

	if terratestConfig.PSACT == string(config.RancherPrivileged) || terratestConfig.PSACT == string(config.RancherRestricted) {
		require.NotEmpty(t, cluster.DefaultPodSecurityAdmissionConfigurationTemplateName)

		err := psact.CreateNginxDeployment(client, clusterID, terratestConfig.PSACT)
		require.NoError(t, err)
	}

	podErrors := pods.StatusPods(client, cluster.ID)
	assert.Empty(t, podErrors)
}

// VerifyNodeCount validates that a cluster has the expected number of nodes.
func VerifyNodeCount(t *testing.T, client *rancher.Client, clusterName string, terraformConfig *config.TerraformConfig, nodeCount int64) {
	clusterID, err := clusterExtensions.GetClusterIDByName(client, clusterName)
	require.NoError(t, err)

	cluster, err := client.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	var module string

	if strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
		module = clustertypes.RKE1
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		module = clustertypes.RKE2
	} else if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
		module = clustertypes.K3S
	} else {
		module = terraformConfig.Module
	}

	switch module {
	case clustertypes.AKS:
		var aksConfig = cluster.AKSConfig
		require.Equal(t, *aksConfig.NodePools[0].Count, cluster.NodeCount)
	case clustertypes.EKS:
		var eksConfig = cluster.EKSConfig
		require.Equal(t, *eksConfig.NodeGroups[0].DesiredSize, cluster.NodeCount)
	case clustertypes.GKE:
		var gkeConfig = cluster.GKEConfig
		require.Equal(t, *gkeConfig.NodePools[0].InitialNodeCount, cluster.NodeCount)
	case clustertypes.RKE1, clustertypes.RKE2, clustertypes.K3S:
		require.Equal(t, nodeCount, cluster.NodeCount)
	default:
		logrus.Errorf("Unsupported module: %v", module)
	}
}

// checkExpectedKubernetesVersion is a helper function that verifies the expected Kubernetes version.
func checkExpectedKubernetesVersion(t *testing.T, terratestConfig *config.TerratestConfig, expectedKubernetesVersion, module string) string {
	switch {
	case module == clustertypes.AKS || module == clustertypes.GKE:
		expectedKubernetesVersion = `v` + terratestConfig.KubernetesVersion
	// Terraform requires that we input the entire RKE1 version. However, Rancher client clips the `-rancher` suffix.
	case strings.Contains(module, clustertypes.RKE1):
		expectedKubernetesVersion = terratestConfig.KubernetesVersion[:len(terratestConfig.KubernetesVersion)-11]
		require.Equal(t, expectedKubernetesVersion, terratestConfig.KubernetesVersion[:len(terratestConfig.KubernetesVersion)-11])
	case strings.Contains(module, clustertypes.EKS) || strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S):
		expectedKubernetesVersion = terratestConfig.KubernetesVersion
		require.Equal(t, expectedKubernetesVersion, terratestConfig.KubernetesVersion)
	default:
		logrus.Errorf("Invalid module provided")
	}

	return expectedKubernetesVersion
}
