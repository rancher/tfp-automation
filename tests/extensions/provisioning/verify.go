package provisioning

import (
	"strings"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	clusterActions "github.com/rancher/tests/actions/clusters"
	"github.com/rancher/tests/actions/psact"
	"github.com/rancher/tests/actions/registries"
	"github.com/rancher/tests/actions/workloads/cronjob"
	"github.com/rancher/tests/actions/workloads/daemonset"
	"github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tests/actions/workloads/statefulset"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	waitState "github.com/rancher/tfp-automation/framework/wait/state"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// VerifyClustersState validates that all clusters are active and have no pod errors.
func VerifyClustersState(t *testing.T, client *rancher.Client, clusterIDs []string) {
	for _, clusterID := range clusterIDs {
		cluster, err := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		logrus.Infof("Waiting for cluster %v to be in an active state...", cluster.Name)
		if err := waitState.IsActiveCluster(client, clusterID); err != nil {
			require.NoError(t, err)
		}

		if err := waitState.AreNodesActive(client, clusterID); err != nil {
			require.NoError(t, err)
		}

		clusterName, err := clusterExtensions.GetClusterNameByID(client, clusterID)
		require.NoError(t, err)

		clusterToken, err := clusterActions.CheckServiceAccountTokenSecret(client, clusterName)
		require.NoError(t, err)
		require.NotEmpty(t, clusterToken)

		podErrors := pods.StatusPods(client, cluster.ID)
		require.Empty(t, podErrors)
	}
}

// VerifyWorkloads validates that different workload operations and workload types are able to provision successfully
func VerifyWorkloads(t *testing.T, client *rancher.Client, clusterIDs []string) {
	workloadValidations := []struct {
		name           string
		validationFunc func(client *rancher.Client, clusterID string) error
	}{
		{"WorkloadDeployment", deployment.VerifyCreateDeployment},
		{"WorkloadSideKick", deployment.VerifyCreateDeploymentSideKick},
		{"WorkloadDaemonSet", daemonset.VerifyCreateDaemonSet},
		{"WorkloadCronjob", cronjob.VerifyCreateCronjob},
		{"WorkloadStatefulset", statefulset.VerifyCreateStatefulset},
		{"WorkloadUpgrade", deployment.VerifyDeploymentUpgradeRollback},
		{"WorkloadPodScaleUp", deployment.VerifyDeploymentPodScaleUp},
		{"WorkloadPodScaleDown", deployment.VerifyDeploymentPodScaleDown},
		{"WorkloadPauseOrchestration", deployment.VerifyDeploymentPauseOrchestration},
	}

	for _, clusterID := range clusterIDs {
		clusterName, err := clusterExtensions.GetClusterNameByID(client, clusterID)
		require.NoError(t, err)

		logrus.Infof("Validating workloads (%s)", clusterName)
		for _, workloadValidation := range workloadValidations {
			retries := 3
			for i := 0; i+1 < retries; i++ {
				err = workloadValidation.validationFunc(client, clusterID)
				if err != nil {
					logrus.Info(err)
					logrus.Infof("Retry %v / %v", i+1, retries)
					continue
				}

				break
			}
			require.NoError(t, err)
		}
	}
}

// VerifyClusterPSACT validates that psact clusters can provision an nginx deployment
func VerifyClusterPSACT(t *testing.T, client *rancher.Client, clusterIDs []string) {
	for _, clusterID := range clusterIDs {
		cluster, err := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		psactName := cluster.DefaultPodSecurityAdmissionConfigurationTemplateName
		if psactName == string(config.RancherPrivileged) || psactName == string(config.RancherRestricted) {
			err := psact.CreateNginxDeployment(client, clusterID, psactName)
			require.NoError(t, err)
		}
	}
}

// VerifyKubernetesVersion validates the expected Kubernetes version.
func VerifyKubernetesVersion(t *testing.T, client *rancher.Client, clusterID, expectedKubernetesVersion, module string) {
	cluster, err := client.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	switch {
	case module == clustertypes.AKS || module == clustertypes.GKE:
		expectedKubernetesVersion = `v` + expectedKubernetesVersion
		require.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)
	case strings.Contains(module, clustertypes.EKS):
		require.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion[1:5])
	case strings.Contains(module, clustertypes.RKE1):
		expectedKubernetesVersion = expectedKubernetesVersion[:len(expectedKubernetesVersion)-11]
		require.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)
	case strings.Contains(module, clustertypes.RKE2) || strings.Contains(module, clustertypes.K3S):
		require.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)
	default:
		logrus.Errorf("Invalid module provided")
	}
}

// VerifyRegistry validates that the expected registry is set.
func VerifyRegistry(t *testing.T, client *rancher.Client, clusterID string, terraformConfig *config.TerraformConfig) {
	if terraformConfig.PrivateRegistries != nil {
		_, err := registries.CheckAllClusterPodsForRegistryPrefix(client, clusterID, terraformConfig.PrivateRegistries.URL)
		require.NoError(t, err)
	}
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
		require.Equal(t, (*aksConfig.NodePools)[0].Count, cluster.NodeCount)
	case clustertypes.EKS:
		var eksConfig = cluster.EKSConfig
		require.Equal(t, (*eksConfig.NodeGroups)[0].DesiredSize, cluster.NodeCount)
	case clustertypes.GKE:
		var gkeConfig = cluster.GKEConfig
		require.Equal(t, (*gkeConfig.NodePools)[0].InitialNodeCount, cluster.NodeCount)
	case clustertypes.RKE1, clustertypes.RKE2, clustertypes.K3S:
		require.Equal(t, nodeCount, cluster.NodeCount)
	default:
		logrus.Errorf("Unsupported module: %v", module)
	}
}
