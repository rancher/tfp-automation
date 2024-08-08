package snapshot

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	apisV1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/shepherd/extensions/etcdsnapshot"
	"github.com/rancher/shepherd/extensions/provisioning"
	"github.com/rancher/shepherd/extensions/services"
	"github.com/rancher/shepherd/extensions/workloads"
	deploy "github.com/rancher/shepherd/extensions/workloads/deployment"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	all                 = "all"
	containerImage      = "nginx"
	containerName       = "nginx"
	defaultNamespace    = "default"
	DeploymentSteveType = "apps.deployment"
	initialWorkload     = "wload-before-restore"
	isCattleLabeled     = true
	kubernetesVersion   = "kubernetesVersion"
	namespace           = "fleet-default"
	port                = "port"
	postWorkload        = "wload-after-backup"
	serviceAppendName   = "service-"
	serviceType         = "service"
)

func snapshotRestore(t *testing.T, client *rancher.Client, clusterName, poolName string, clusterConfig *config.TerratestConfig, terraformOptions *terraform.Options) {
	initialWorkloadName := namegen.AppendRandomString(initialWorkload)

	clusterID, err := clusters.GetClusterIDByName(client, clusterName)
	require.NoError(t, err)

	steveclient, err := client.Steve.ProxyDownstream(clusterID)
	require.NoError(t, err)

	localClusterID, err := clusters.GetClusterIDByName(client, clustertypes.Local)
	require.NoError(t, err)

	containerTemplate := workloads.NewContainer(containerName, containerImage, corev1.PullAlways, []corev1.VolumeMount{}, []corev1.EnvFromSource{}, nil, nil, nil)
	podTemplate := workloads.NewPodTemplate([]corev1.Container{containerTemplate}, []corev1.Volume{}, []corev1.LocalObjectReference{}, nil)
	deployment := workloads.NewDeploymentTemplate(initialWorkloadName, defaultNamespace, podTemplate, isCattleLabeled, nil)

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAppendName + initialWorkloadName,
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

	deploymentResp, err := deploy.CreateDeployment(steveclient, initialWorkloadName, deployment)
	require.NoError(t, err)

	err = deploy.VerifyDeployment(steveclient, deploymentResp)
	require.NoError(t, err)
	require.Equal(t, initialWorkloadName, deploymentResp.ObjectMeta.Name)

	serviceResp, err := services.CreateService(steveclient, service)
	require.NoError(t, err)

	err = services.VerifyService(steveclient, serviceResp)
	require.NoError(t, err)
	require.Equal(t, serviceAppendName+initialWorkloadName, serviceResp.ObjectMeta.Name)

	cluster, snapshotName, postDeploymentResp, postServiceResp := snapshotV2Prov(t, client, podTemplate, deployment, clusterName, poolName, clusterID, localClusterID, clusterConfig, false, terraformOptions)
	restoreV2Prov(t, client, clusterConfig, snapshotName, clusterName, poolName, cluster, clusterID, terraformOptions)

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

func snapshotV2Prov(t *testing.T, client *rancher.Client, podTemplate corev1.PodTemplateSpec, deployment *v1.Deployment, clusterName, poolName, clusterID, localClusterID string,
	clusterConfig *config.TerratestConfig, isRKE1 bool, terraformOptions *terraform.Options) (*apisV1.Cluster, string, *steveV1.SteveAPIObject, *steveV1.SteveAPIObject) {
	existingSnapshots, err := etcdsnapshot.GetRKE2K3SSnapshots(client, clusterName)
	require.NoError(t, err)

	clusterConfig.SnapshotInput.CreateSnapshot = true

	err = framework.ConfigTF(nil, clusterConfig, clusterName, poolName, "")
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)

	err = clusters.WaitClusterToBeUpgraded(client, clusterID)
	require.NoError(t, err)

	cluster, _, err := clusters.GetProvisioningClusterByName(client, clusterName, namespace)
	require.NoError(t, err)

	podErrors := pods.StatusPods(client, clusterID)
	assert.Empty(t, podErrors)

	postDeploymentResp, postServiceResp := createPostBackupWorkloads(t, client, clusterID, podTemplate, deployment)

	etcdNodeCount, _ := etcdsnapshot.MatchNodeToAnyEtcdRole(client, clusterID)
	snapshotToRestore, err := provisioning.VerifySnapshots(client, clusterName, etcdNodeCount+len(existingSnapshots), isRKE1)
	require.NoError(t, err)

	if clusterConfig.SnapshotInput.SnapshotRestore == kubernetesVersion || clusterConfig.SnapshotInput.SnapshotRestore == all {
		clusterObject, _, err := clusters.GetProvisioningClusterByName(client, clusterName, namespace)
		require.NoError(t, err)

		initialKubernetesVersion := clusterObject.Spec.KubernetesVersion

		if clusterConfig.SnapshotInput.UpgradeKubernetesVersion == "" {
			if strings.Contains(initialKubernetesVersion, clustertypes.RKE2) {
				defaultVersion, err := kubernetesversions.Default(client, clusters.RKE2ClusterType.String(), nil)
				clusterConfig.SnapshotInput.UpgradeKubernetesVersion = defaultVersion[0]
				require.NoError(t, err)
			} else if strings.Contains(initialKubernetesVersion, clustertypes.K3S) {
				defaultVersion, err := kubernetesversions.Default(client, clusters.K3SClusterType.String(), nil)
				clusterConfig.SnapshotInput.UpgradeKubernetesVersion = defaultVersion[0]
				require.NoError(t, err)
			}
		}

		clusterObject.Spec.KubernetesVersion = clusterConfig.SnapshotInput.UpgradeKubernetesVersion

		if clusterConfig.SnapshotInput.SnapshotRestore == all && clusterConfig.SnapshotInput.ControlPlaneConcurrencyValue != "" && clusterConfig.SnapshotInput.WorkerConcurrencyValue != "" {
			clusterObject.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency = clusterConfig.SnapshotInput.ControlPlaneConcurrencyValue
			clusterObject.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency = clusterConfig.SnapshotInput.WorkerConcurrencyValue
		}

		clusterConfig.KubernetesVersion = clusterObject.Spec.KubernetesVersion
		clusterConfig.SnapshotInput.CreateSnapshot = false

		err = framework.ConfigTF(nil, clusterConfig, clusterName, poolName, "")
		require.NoError(t, err)

		terraform.Apply(t, terraformOptions)

		err = clusters.WaitClusterToBeUpgraded(client, clusterID)
		require.NoError(t, err)

		logrus.Infof("Cluster version is upgraded to: %s", clusterObject.Spec.KubernetesVersion)

		podErrors := pods.StatusPods(client, clusterID)
		assert.Empty(t, podErrors)
		require.Equal(t, clusterConfig.SnapshotInput.UpgradeKubernetesVersion, clusterObject.Spec.KubernetesVersion)

		if clusterConfig.SnapshotInput.SnapshotRestore == all && clusterConfig.SnapshotInput.ControlPlaneConcurrencyValue != "" && clusterConfig.SnapshotInput.WorkerConcurrencyValue != "" {
			logrus.Infof("Control plane concurrency value is set to: %s", clusterObject.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency)
			logrus.Infof("Worker concurrency value is set to: %s", clusterObject.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency)

			require.Equal(t, clusterConfig.SnapshotInput.ControlPlaneConcurrencyValue, clusterObject.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency)
			require.Equal(t, clusterConfig.SnapshotInput.WorkerConcurrencyValue, clusterObject.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency)
		}
	}

	return cluster, snapshotToRestore, postDeploymentResp, postServiceResp
}

func restoreV2Prov(t *testing.T, client *rancher.Client, clusterConfig *config.TerratestConfig, snapshotName, clusterName, poolName string, cluster *apisV1.Cluster, clusterID string, terraformOptions *terraform.Options) {
	clusterConfig.SnapshotInput.CreateSnapshot = false
	clusterConfig.SnapshotInput.RestoreSnapshot = true
	clusterConfig.SnapshotInput.SnapshotName = snapshotName

	err := framework.ConfigTF(nil, clusterConfig, clusterName, poolName, "")
	require.NoError(t, err)

	terraform.Apply(t, terraformOptions)

	err = clusters.WaitClusterToBeUpgraded(client, clusterID)
	require.NoError(t, err)

	clusterObject, _, err := clusters.GetProvisioningClusterByName(client, clusterName, namespace)
	require.NoError(t, err)

	logrus.Infof("Cluster version is restored to: %s", clusterObject.Spec.KubernetesVersion)

	podErrors := pods.StatusPods(client, clusterID)
	assert.Empty(t, podErrors)

	if clusterConfig.SnapshotInput.SnapshotRestore == kubernetesVersion || clusterConfig.SnapshotInput.SnapshotRestore == all {
		clusterObject, _, err := clusters.GetProvisioningClusterByName(client, clusterName, namespace)
		require.NoError(t, err)
		require.Equal(t, cluster.Spec.KubernetesVersion, clusterObject.Spec.KubernetesVersion)

		if clusterConfig.SnapshotInput.ControlPlaneConcurrencyValue != "" && clusterConfig.SnapshotInput.WorkerConcurrencyValue != "" {
			logrus.Infof("Control plane concurrency value is restored to: %s", clusterObject.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency)
			logrus.Infof("Worker concurrency value is restored to: %s", clusterObject.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency)

			require.Equal(t, cluster.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency, clusterObject.Spec.RKEConfig.UpgradeStrategy.ControlPlaneConcurrency)
			require.Equal(t, cluster.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency, clusterObject.Spec.RKEConfig.UpgradeStrategy.WorkerConcurrency)
		}
	}

}

func createPostBackupWorkloads(t *testing.T, client *rancher.Client, clusterID string, podTemplate corev1.PodTemplateSpec, deployment *v1.Deployment) (*steveV1.SteveAPIObject, *steveV1.SteveAPIObject) {
	workloadNamePostBackup := namegen.AppendRandomString(postWorkload)

	postBackupDeployment := workloads.NewDeploymentTemplate(workloadNamePostBackup, defaultNamespace, podTemplate, isCattleLabeled, nil)
	postBackupService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAppendName + workloadNamePostBackup,
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

	postDeploymentResp, err := deploy.CreateDeployment(steveclient, workloadNamePostBackup, postBackupDeployment)
	require.NoError(t, err)

	err = deploy.VerifyDeployment(steveclient, postDeploymentResp)
	require.NoError(t, err)
	require.Equal(t, workloadNamePostBackup, postDeploymentResp.ObjectMeta.Name)

	postServiceResp, err := services.CreateService(steveclient, postBackupService)
	require.NoError(t, err)

	err = services.VerifyService(steveclient, postServiceResp)
	require.NoError(t, err)
	require.Equal(t, serviceAppendName+workloadNamePostBackup, postServiceResp.ObjectMeta.Name)

	return postDeploymentResp, postServiceResp
}
