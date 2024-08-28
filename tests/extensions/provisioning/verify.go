package provisioning

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	clusterActions "github.com/rancher/rancher/tests/v2/actions/clusters"
	"github.com/rancher/rancher/tests/v2/actions/psact"
	"github.com/rancher/shepherd/clients/rancher"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	waitState "github.com/rancher/tfp-automation/framework/wait/state"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// VerifyCluster validates that a downstream cluster and its resources are in a good state, matching a given config.
func VerifyCluster(t *testing.T, client *rancher.Client, clusterName string, terraformConfig *config.TerraformConfig, terraformOptions *terraform.Options, clusterConfig *config.TerratestConfig) {
	client, err := client.ReLogin()
	require.NoError(t, err)

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	require.NoError(t, err)

	var expectedKubernetesVersion string
	module := terraformConfig.Module
	expectedKubernetesVersion = checkExpectedKubernetesVersion(clusterConfig, expectedKubernetesVersion, module)

	clusterID, err := clusterExtensions.GetClusterIDByName(adminClient, clusterName)
	require.NoError(t, err)

	if err := waitState.IsActiveCluster(adminClient, clusterID); err != nil {
		require.NoError(t, err)
	}

	if err := waitState.AreNodesActive(adminClient, clusterID); err != nil {
		require.NoError(t, err)
	}

	cluster, err := adminClient.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	// EKS is formatted this way due to EKS formatting Kubernetes versions with a
	// random string of letters after the version.
	if module == clustertypes.EKS {
		assert.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion[1:5])
	} else {
		assert.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)
	}

	clusterToken, err := clusterActions.CheckServiceAccountTokenSecret(adminClient, cluster.Name)
	require.NoError(t, err)
	assert.NotEmpty(t, clusterToken)

	if clusterConfig.PSACT == string(config.RancherPrivileged) || clusterConfig.PSACT == string(config.RancherRestricted) {
		require.NotEmpty(t, cluster.DefaultPodSecurityAdmissionConfigurationTemplateName)

		err := psact.CreateNginxDeployment(adminClient, clusterID, clusterConfig.PSACT)
		require.NoError(t, err)
	}

	podErrors := pods.StatusPods(adminClient, cluster.ID)
	assert.Empty(t, podErrors)
}

// VerifyUpgradedKubernetesVersion validates that the cluster has the expected
// upgraded Kubernetes version.
func VerifyUpgradedKubernetesVersion(t *testing.T, client *rancher.Client, terraformConfig *config.TerraformConfig, clusterName, kubernetesVersion string) {
	client, err := client.ReLogin()
	require.NoError(t, err)

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	require.NoError(t, err)

	clusterID, err := clusterExtensions.GetClusterIDByName(adminClient, clusterName)
	require.NoError(t, err)

	cluster, err := adminClient.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	if cluster.Version.GitVersion == kubernetesVersion {
		logrus.Infof("Successfully upgraded cluster to %v!", kubernetesVersion)
		require.NoError(t, err)
	}
}

// VerifyNodeCount validates that a cluster has the expected number of nodes.
func VerifyNodeCount(t *testing.T, client *rancher.Client, clusterName string, terraformConfig *config.TerraformConfig, nodeCount int64) {
	client, err := client.ReLogin()
	require.NoError(t, err)

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	require.NoError(t, err)

	clusterID, err := clusterExtensions.GetClusterIDByName(adminClient, clusterName)
	require.NoError(t, err)

	err = kwait.PollUntilContextTimeout(context.TODO(), 10*time.Second, defaults.FifteenMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		cluster, err := adminClient.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		if strings.Contains(terraformConfig.Module, clustertypes.RKE1) {
			terraformConfig.Module = clustertypes.RKE1
		} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
			terraformConfig.Module = clustertypes.RKE2
		} else if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
			terraformConfig.Module = clustertypes.K3S
		}

		switch terraformConfig.Module {
		case clustertypes.AKS:
			var aksConfig = cluster.AKSConfig
			if cluster.NodeCount == *aksConfig.NodePools[0].Count {
				logrus.Infof("Successfully scaled cluster to %v total nodes!", *aksConfig.NodePools[0].Count)
			}

			return true, nil
		case clustertypes.EKS:
			var eksConfig = cluster.EKSConfig
			if cluster.NodeCount == *eksConfig.NodeGroups[0].DesiredSize {
				logrus.Infof("Successfully scaled cluster to %v total nodes!", *eksConfig.NodeGroups[0].DesiredSize)
			}

			return true, nil
		case clustertypes.GKE:
			var gkeConfig = cluster.GKEConfig
			if cluster.NodeCount == *gkeConfig.NodePools[0].InitialNodeCount {
				logrus.Infof("Successfully scaled cluster to %v total nodes!", *gkeConfig.NodePools[0].InitialNodeCount)
			}

			return true, nil
		case clustertypes.RKE1, clustertypes.RKE2, clustertypes.K3S:
			if cluster.NodeCount == nodeCount {
				logrus.Infof("Successfully scaled cluster to %v total nodes!", nodeCount)

				return true, nil
			}
		default:
			logrus.Errorf("Unsupported module: %v", terraformConfig.Module)
		}

		return false, nil
	})
	require.NoError(t, err)
}

// checkExpectedKubernetesVersion is a helper function that verifies the expected Kubernetes version.
func checkExpectedKubernetesVersion(clusterConfig *config.TerratestConfig, expectedKubernetesVersion, module string) string {
	switch {
	case module == clustertypes.AKS || module == clustertypes.GKE:
		expectedKubernetesVersion = `v` + clusterConfig.KubernetesVersion
	// Terraform requires that we input the entire RKE1 version. However, Rancher client clips the `-rancher` suffix.
	case strings.Contains(module, clustertypes.RKE1):
		expectedKubernetesVersion = clusterConfig.KubernetesVersion[:len(clusterConfig.KubernetesVersion)-11]
	case strings.Contains(module, clustertypes.EKS) || strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S):
		expectedKubernetesVersion = clusterConfig.KubernetesVersion
	default:
		logrus.Errorf("Invalid module provided")
	}

	return expectedKubernetesVersion
}
