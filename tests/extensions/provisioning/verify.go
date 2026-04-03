package provisioning

import (
	"context"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	clusterActions "github.com/rancher/tests/actions/clusters"
	"github.com/rancher/tests/actions/psact"
	"github.com/rancher/tests/actions/registries"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/tests/extensions/postProvisioning"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

// VerifyClusterReady validates that a cluster is active and ready
func VerifyClusterReady(client *rancher.Client, cluster *steveV1.SteveAPIObject) error {
	status := &provv1.ClusterStatus{}
	err := steveV1.ConvertToK8sType(cluster.Status, status)
	if err != nil {
		return err
	}

	return postProvisioning.IsClusterActive(client, status.ClusterName)
}

// VerifyV3ClustersPods validates that all pods in the v3 clusters are running without errors.
func VerifyV3ClustersPods(t *testing.T, client *rancher.Client, clusterIDs []string) {
	for _, clusterID := range clusterIDs {
		podErrors := pods.StatusPods(client, clusterID)
		require.Empty(t, podErrors)
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

	logrus.Infof("Rancher version: %s | Rancher commit: %s", resp.RancherVersion, resp.GitCommit)

	if resp.RancherVersion != expectedVersion {
		logrus.Infof("Expected version: %s | Actual version: %s", expectedVersion, resp.RancherVersion)
		cleanup.Cleanup(t, terraformOptions, keyPath)
	}
}

// VerifyACEAirgap validates that the ACE resources are healthy in a given airgap cluster
func VerifyACEAirgap(t *testing.T, client *rancher.Client, cluster *steveV1.SteveAPIObject) {
	status := &provv1.ClusterStatus{}
	err := steveV1.ConvertToK8sType(cluster.Status, status)
	require.NoError(t, err)

	clusterObject, err := client.Management.Cluster.ByID(status.ClusterName)
	require.NoError(t, err)

	kubeConfig, err := kubeconfig.GetKubeconfig(client, clusterObject.ID)
	require.NoError(t, err)

	clientConfig := *kubeConfig

	rawConfig, err := clientConfig.RawConfig()
	require.NoError(t, err)

	var contextNames []string
	for name := range rawConfig.Contexts {
		if strings.Contains(name, "pool") {
			contextNames = append(contextNames, name)
		}
	}

	for _, contextName := range contextNames {
		restConfig, err := clientcmd.NewNonInteractiveClientConfig(rawConfig, contextName, &clientcmd.ConfigOverrides{}, nil).ClientConfig()
		require.NoError(t, err)

		k8sClient, err := kubernetes.NewForConfig(restConfig)
		require.NoError(t, err)

		pods, err := k8sClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		require.NoError(t, err)

		logrus.Infof("Switched context to %v", contextName)
		for _, pod := range pods.Items {
			logrus.Debugf("Pod %v", pod.GetName())
		}
	}
}
