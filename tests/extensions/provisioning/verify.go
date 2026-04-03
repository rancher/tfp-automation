package provisioning

import (
	"context"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/tests/actions/registries"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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
