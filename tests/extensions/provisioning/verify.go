package provisioning

import (
	"strings"
	"testing"

	apisV1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	clusterActions "github.com/rancher/rancher/tests/v2/actions/clusters"
	"github.com/rancher/rancher/tests/v2/actions/psact"
	"github.com/rancher/rancher/tests/v2/actions/registries"
	"github.com/rancher/rancher/tests/v2/actions/workloads/cronjob"
	"github.com/rancher/rancher/tests/v2/actions/workloads/daemonset"
	"github.com/rancher/rancher/tests/v2/actions/workloads/deployment"
	"github.com/rancher/rancher/tests/v2/actions/workloads/statefulset"
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
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

		v1ClusterID, err := clusterExtensions.GetV1ProvisioningClusterByName(client, clusterName)
		require.NoError(t, err)

		var v1Cluster *v1.SteveAPIObject
		if v1ClusterID == "" {
			v1Cluster, err = client.Steve.SteveType(stevetypes.Provisioning).ByID("fleet-default/" + clusterID)
			require.NoError(t, err)
			require.NotEmpty(t, v1Cluster)
		} else {
			v1Cluster, err = client.Steve.SteveType(stevetypes.Provisioning).ByID(v1ClusterID)
			require.NoError(t, err)
			require.NotEmpty(t, v1Cluster)
		}

		clusterObj := new(apisV1.Cluster)
		err = v1.ConvertToK8sType(v1Cluster, &clusterObj)
		require.NoError(t, err)

		if clusterObj.Spec.RKEConfig != nil {
			if clusterObj.Spec.RKEConfig.Registries != nil {
				for registryURL := range clusterObj.Spec.RKEConfig.Registries.Configs {
					_, err := registries.CheckAllClusterPodsForRegistryPrefix(client, clusterID, registryURL)
					require.NoError(t, err)
				}
			}
		}
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

	// Terraform requires that we input the entire RKE1 version. However, Rancher client clips the `-rancher` suffix.
	case strings.Contains(module, clustertypes.RKE1):
		expectedKubernetesVersion = expectedKubernetesVersion[:len(expectedKubernetesVersion)-11]
		require.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)

	case strings.Contains(module, clustertypes.EKS):
		require.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion[1:5])

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
