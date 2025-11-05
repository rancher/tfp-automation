package provisioning

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
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
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/postProvisioning"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// VerifyClustersState validates that all clusters are active and have no pod errors.
func VerifyClustersState(t *testing.T, client *rancher.Client, clusterIDs []string) {
	for _, clusterID := range clusterIDs {
		cluster, err := client.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		logrus.Infof("Waiting for cluster %v to be in an active state...", cluster.Name)
		err = postProvisioning.IsClusterActive(client, clusterID)
		require.NoError(t, err)

		logrus.Infof("Waiting for all nodes to be active on cluster: %s", cluster.Name)
		err = postProvisioning.AreNodesActive(client, clusterID)
		require.NoError(t, err)
	}
}

// VerifyV3ClustersPods validates that all pods in the v3 clusters are running without errors.
func VerifyV3ClustersPods(t *testing.T, client *rancher.Client, clusterIDs []string) {
	for _, clusterID := range clusterIDs {
		podErrors := pods.StatusPods(client, clusterID)
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

// VerifyServiceAccountTokenSecret validates that the service account token secret exists for each cluster
func VerifyServiceAccountTokenSecret(t *testing.T, client *rancher.Client, clusterIDs []string) {
	for _, clusterID := range clusterIDs {
		clusterName, err := clusterExtensions.GetClusterNameByID(client, clusterID)
		require.NoError(t, err)

		clusterToken, err := clusterActions.CheckServiceAccountTokenSecret(client, clusterName)
		require.NoError(t, err)
		require.NotEmpty(t, clusterToken)
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

// VerifyRegistry validates that the expected registry is set.
func VerifyRegistry(t *testing.T, client *rancher.Client, clusterID string, terraformConfig *config.TerraformConfig) {
	if terraformConfig.PrivateRegistries != nil {
		_, err := registries.CheckAllClusterPodsForRegistryPrefix(client, clusterID, terraformConfig.PrivateRegistries.URL)
		require.NoError(t, err)
	}
}

// VerifyRancherVersion validates that the expected rancher version matches the version of the rancher server.
func VerifyRancherVersion(t *testing.T, hostURL, expectedVersion, keyPath string, terraformOptions *terraform.Options) {
	resp, err := RequestRancherVersion(hostURL)
	require.NoError(t, err)

	if resp.RancherVersion != expectedVersion {
		logrus.Infof("Expected version: %s | Actual version: %s", expectedVersion, resp.RancherVersion)
		cleanup.Cleanup(t, terraformOptions, keyPath)
	}
}
