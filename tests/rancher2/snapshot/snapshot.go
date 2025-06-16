package snapshot

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/clusters"
	timeouts "github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/workloads"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"github.com/rancher/shepherd/pkg/config/operations"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tests/actions/services"
	deploy "github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	StorageAnnotation        = "etcdsnapshot.rke.io/storage"
	SnapshotAnnotation       = "rke.cattle.io.etcdsnapshot"
	SnapshotClusterNameLabel = "rke.cattle.io/cluster-name"

	active              = "active"
	all                 = "all"
	containerImage      = "nginx"
	containerName       = "nginx"
	defaultNamespace    = "default"
	DeploymentSteveType = "apps.deployment"
	initialWorkload     = "wload-before-restore"
	isCattleLabeled     = true
	localCluster        = "local"
	kubernetesVersion   = "kubernetesVersion"
	namespace           = "fleet-default"
	port                = "port"
	postWorkload        = "wload-after-backup"
	S3                  = "s3"
	serviceAppendName   = "service-"
	serviceType         = "service"
)

// RestoreSnapshot creates workloads, takes a snapshot of the cluster, restores the cluster and verifies the workloads created after
// a snapshot no longer are present in the cluster
func RestoreSnapshot(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, testUser, testPassword string, terraformOptions *terraform.Options, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) {
	initialWorkloadName := namegen.AppendRandomString(initialWorkload)

	clusterID, err := clusters.GetClusterIDByName(client, terraformConfig.ResourcePrefix)
	require.NoError(t, err)

	steveclient, err := client.Steve.ProxyDownstream(clusterID)
	require.NoError(t, err)

	containerTemplate := workloads.NewContainer(containerName, containerImage, corev1.PullAlways, []corev1.VolumeMount{}, []corev1.EnvFromSource{}, nil, nil, nil)
	podTemplate := workloads.NewPodTemplate([]corev1.Container{containerTemplate}, []corev1.Volume{}, []corev1.LocalObjectReference{}, nil, nil)

	deploymentResp, serviceResp := createWorkloads(t, client, clusterID, podTemplate, initialWorkloadName, isCattleLabeled, DeploymentSteveType)

	snapshotName, postDeploymentResp, postServiceResp, err := snapshotV2Prov(t, client, rancherConfig, terraformConfig, terratestConfig, podTemplate, testUser, testPassword, clusterID, terraformOptions, configMap, newFile, rootBody, file)
	require.NoError(t, err)

	restoreV2Prov(t, client, rancherConfig, terraformConfig, terratestConfig, snapshotName, testUser, testPassword, clusterID, terraformOptions, configMap, newFile, rootBody, file)

	_, err = steveclient.SteveType(DeploymentSteveType).ByID(postDeploymentResp.ID)
	require.Error(t, err)

	_, err = steveclient.SteveType(serviceType).ByID(postServiceResp.ID)
	require.Error(t, err)

	logrus.Infof("Deleting created workloads...")
	err = steveclient.SteveType(stevetypes.Deployment).Delete(deploymentResp)
	require.NoError(t, err)

	err = steveclient.SteveType(stevetypes.Service).Delete(serviceResp)
	require.NoError(t, err)
}

// snapshotV2Prov takes a snapshot of the cluster and creates a deployment and service in the cluster.
func snapshotV2Prov(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, podTemplate corev1.PodTemplateSpec, testUser, testPassword, clusterID string,
	terraformOptions *terraform.Options, configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File) (string, *steveV1.SteveAPIObject, *steveV1.SteveAPIObject, error) {
	_, err := operations.ReplaceValue([]string{"terratest", "snapshotInput", "createSnapshot"}, true, configMap[0])
	require.NoError(t, err)

	_, _, err = framework.ConfigTF(client, rancherConfig, terratestConfig, testUser, testPassword, "", configMap, newFile, rootBody, file, false, false, false, nil)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)

	err = clusters.WaitClusterToBeUpgraded(client, clusterID)
	require.NoError(t, err)

	podErrors := pods.StatusPods(client, clusterID)
	assert.Empty(t, podErrors)

	postWorkloadName := namegen.AppendRandomString(postWorkload)
	postDeploymentResp, postServiceResp := createWorkloads(t, client, clusterID, podTemplate, postWorkloadName, isCattleLabeled, DeploymentSteveType)

	snapshotID, err := getSnapshots(client, terraformConfig.ResourcePrefix)
	require.NoError(t, err)

	return snapshotID[0].Name, postDeploymentResp, postServiceResp, err
}

// restoreV2Prov restores the cluster to the previous state after a snapshot is taken.
func restoreV2Prov(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, snapshotName, testUser, testPassword string, clusterID string, terraformOptions *terraform.Options,
	configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) {
	_, err := operations.ReplaceValue([]string{"terratest", "snapshotInput", "createSnapshot"}, false, configMap[0])
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"terratest", "snapshotInput", "restoreSnapshot"}, true, configMap[0])
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"terratest", "snapshotInput", "snapshotName"}, snapshotName, configMap[0])
	require.NoError(t, err)

	_, _, err = framework.ConfigTF(client, rancherConfig, terratestConfig, testUser, testPassword, "", configMap, newFile, rootBody, file, false, false, false, nil)
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)

	err = clusters.WaitClusterToBeUpgraded(client, clusterID)
	require.NoError(t, err)

	clusterObject, _, err := clusters.GetProvisioningClusterByName(client, terraformConfig.ResourcePrefix, namespace)
	require.NoError(t, err)

	logrus.Infof("Cluster version is restored to: %s", clusterObject.Spec.KubernetesVersion)

	podErrors := pods.StatusPods(client, clusterID)
	assert.Empty(t, podErrors)
}

// getSnapshots retrieves all snapshots for a given cluster.
func getSnapshots(client *rancher.Client, clusterName string) ([]steveV1.SteveAPIObject, error) {
	localclusterID, err := clusters.GetClusterIDByName(client, localCluster)
	if err != nil {
		return nil, err
	}

	steveclient, err := client.Steve.ProxyDownstream(localclusterID)
	if err != nil {
		return nil, err
	}

	snapshotSteveObjList, err := steveclient.SteveType(SnapshotAnnotation).List(nil)
	if err != nil {
		return nil, err
	}

	snapshots := []steveV1.SteveAPIObject{}
	for _, snapshot := range snapshotSteveObjList.Data {
		if strings.Contains(snapshot.ObjectMeta.Name, clusterName) {
			snapshots = append(snapshots, snapshot)
		}
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].ObjectMeta.CreationTimestamp.Before(&snapshots[j].ObjectMeta.CreationTimestamp)
	})

	return snapshots, nil
}

// createWorkloads creates a deployment and service in a given cluster and verifies they are active.
func createWorkloads(t *testing.T, client *rancher.Client, clusterID string, podTemplate corev1.PodTemplateSpec, workloadName string, isCattleLabeled bool, deploymentType string) (*steveV1.SteveAPIObject, *steveV1.SteveAPIObject) {
	deployment := workloads.NewDeploymentTemplate(workloadName, defaultNamespace, podTemplate, isCattleLabeled, nil)

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAppendName + workloadName,
			Namespace: defaultNamespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: port,
					Port: 80,
				},
			},
			Selector: deployment.Spec.Template.Labels,
		},
	}

	steveclient, err := client.Steve.ProxyDownstream(clusterID)
	require.NoError(t, err)

	deploymentResp, err := steveclient.SteveType(deploymentType).Create(deployment)
	require.NoError(t, err)

	err = kwait.PollUntilContextTimeout(context.TODO(), timeouts.FiveSecondTimeout, timeouts.FiveMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		deployment, err := client.Steve.SteveType(deploymentType).ByID(deploymentResp.ID)
		if err != nil {
			return false, err
		}

		if deployment.State.Name == active {
			logrus.Infof("%s(%s) is active", deploymentType, deployment.Name)
			return true, nil
		}

		return false, nil
	})

	err = deploy.VerifyDeployment(steveclient, deploymentResp)
	require.NoError(t, err)
	require.Equal(t, workloadName, deploymentResp.ObjectMeta.Name)

	serviceResp, err := services.CreateService(steveclient, service)
	require.NoError(t, err)

	err = services.VerifyService(steveclient, serviceResp)
	require.NoError(t, err)
	require.Equal(t, serviceAppendName+workloadName, serviceResp.ObjectMeta.Name)

	return deploymentResp, serviceResp
}
