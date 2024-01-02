package provisioning

import (
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/psact"
	"github.com/josh-diamond/tfp-automation/config"
	waitState "github.com/josh-diamond/tfp-automation/framework/wait/state"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	machineNameAnnotation    = "cluster.x-k8s.io/machine"
	machineSteveResourceType = "cluster.x-k8s.io.machine"
	aks                      = "aks"
	eks                      = "eks"
	gke                      = "gke"
	ec2RKE1                  = "ec2_rke1"
	ec2RKE2                  = "ec2_rke2"
	ec2K3s                   = "ec2_k3s"
	linodeRKE1               = "linode_rke1"
	linodeRKE2               = "linode_rke2"
	linodeK3s                = "linode_k3s"
	rke1                     = "rke1"
	rke2                     = "rke2"
	k3s                      = "k3s"

	active                        = "active"
	ProvisioningSteveResourceType = "provisioning.cattle.io.cluster"
	provisioning                  = "provisioning"
	kubernetesUpgrade             = "kubernetes-upgrade"
	scaleUp                       = "scale-up"
	scaleDown                     = "scale-down"
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

	clusterID, err := clusters.GetClusterIDByName(adminClient, clusterName)
	require.NoError(t, err)

	waitState.IsActiveCluster(adminClient, clusterID)
	waitState.AreNodesActive(adminClient, clusterID)

	cluster, err := adminClient.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	// EKS is formatted this way due to EKS formatting Kubernetes versions with a random string of letters after the version.
	if module == eks {
		assert.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion[1:5])
	} else {
		assert.Equal(t, expectedKubernetesVersion, cluster.Version.GitVersion)
	}

	clusterToken, err := clusters.CheckServiceAccountTokenSecret(adminClient, cluster.Name)
	require.NoError(t, err)
	assert.NotEmpty(t, clusterToken)

	if clusterConfig.PSACT == string(config.RancherPrivileged) || clusterConfig.PSACT == string(config.RancherRestricted) {
		require.NotEmpty(t, cluster.DefaultPodSecurityAdmissionConfigurationTemplateName)

		err := psact.CreateNginxDeployment(adminClient, clusterID, clusterConfig.PSACT)
		require.NoError(t, err)
	}
}

// VerifyUpgradedKubernetesVersion validates that the cluster has the expected upgraded Kubernetes version.
func VerifyUpgradedKubernetesVersion(t *testing.T, client *rancher.Client, terraformConfig *config.TerraformConfig, clusterName, kubernetesVersion string) {
	client, err := client.ReLogin()
	require.NoError(t, err)

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	require.NoError(t, err)

	clusterID, err := clusters.GetClusterIDByName(adminClient, clusterName)
	require.NoError(t, err)

	cluster, err := adminClient.Management.Cluster.ByID(clusterID)
	require.NoError(t, err)

	if cluster.Version.GitVersion == kubernetesVersion {
		logrus.Infof("Successfully upgraded cluster to %v!", kubernetesVersion)
		require.NoError(t, err)
	}
}

// VerifyNodeCount validates that a cluster has the expected number of nodes.
func VerifyNodeCount(t *testing.T, client *rancher.Client, clusterName string, nodeCount int64) {
	client, err := client.ReLogin()
	require.NoError(t, err)

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	require.NoError(t, err)

	clusterID, err := clusters.GetClusterIDByName(adminClient, clusterName)
	require.NoError(t, err)

	err = wait.Poll(10*time.Second, 10*time.Minute, func() (done bool, err error) {
		cluster, err := adminClient.Management.Cluster.ByID(clusterID)
		require.NoError(t, err)

		if cluster.NodeCount == nodeCount {
			logrus.Infof("Successfully scaled cluster to %v total nodes!", nodeCount)
			return true, nil
		}

		return false, nil
	})
	require.NoError(t, err)
}

// checkExpectedKubernetesVersion is a helper function that verifies the expected Kubernetes version.
func checkExpectedKubernetesVersion(clusterConfig *config.TerratestConfig, expectedKubernetesVersion, module string) string {
	switch {
	case module == aks || module == gke:
		expectedKubernetesVersion = `v` + clusterConfig.KubernetesVersion
	// Terraform requires that we input the entire RKE1 version. However, Rancher client clips the `-rancher` suffix.
	case strings.Contains(module, rke1):
		expectedKubernetesVersion = clusterConfig.KubernetesVersion[:len(clusterConfig.KubernetesVersion)-11]
	case strings.Contains(module, eks) || strings.Contains(module, rke2) || strings.Contains(module, k3s):
		expectedKubernetesVersion = clusterConfig.KubernetesVersion
	default:
		logrus.Errorf("Invalid module provided")
	}

	return expectedKubernetesVersion
}
