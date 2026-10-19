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
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	StorageAnnotation        = "etcdsnapshot.rke.io/storage"
	SnapshotAnnotation       = "rke.cattle.io.etcdsnapshot"
	SnapshotClusterNameLabel = "rke.cattle.io/cluster-name"

	DeploymentSteveType = "apps.deployment"
	localCluster        = "local"
	namespace           = "fleet-default"
	port                = "port"
	S3                  = "s3"
)

// RestoreSnapshot creates workloads, takes a snapshot of the cluster, restores the cluster and verifies the workloads created after
// a snapshot no longer are present in the cluster
func RestoreSnapshot(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, terraformOptions *terraform.Options, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File, nestedRancherModuleDir string) {
	clusterID, err := clusters.GetClusterIDByName(client, terraformConfig.ResourcePrefix)
	require.NoError(t, err)

	snapshotName, err := snapshotV2Prov(t, client, rancherConfig, terraformConfig, terratestConfig, clusterID, terraformOptions, newFile, rootBody, file, nestedRancherModuleDir)
	require.NoError(t, err)

	restoreV2Prov(t, client, rancherConfig, terraformConfig, terratestConfig, snapshotName, terraformOptions, newFile, rootBody, file, nestedRancherModuleDir)
}

// snapshotV2Prov takes a snapshot of the cluster and creates a deployment and service in the cluster.
func snapshotV2Prov(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, clusterID string, terraformOptions *terraform.Options, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File, nestedRancherModuleDir string) (string, error) {
	terratestConfig.SnapshotInput.CreateSnapshot = true

	_, _, err := framework.ConfigTF(client, rancherConfig, terratestConfig, "", terraformConfig, newFile, rootBody, file, false, false, false, "", nestedRancherModuleDir)
	require.NoError(t, err)

	terraform.ApplyContext(t, context.Background(), terraformOptions)

	err = clusters.WaitClusterToBeUpgraded(client, clusterID)
	require.NoError(t, err)

	var snapshotID []steveV1.SteveAPIObject
	err = kwait.PollUntilContextTimeout(context.TODO(), timeouts.FiveSecondTimeout, timeouts.FiveMinuteTimeout, true, func(ctx context.Context) (bool, error) {
		snapshotID, err = getSnapshots(client, terraformConfig.ResourcePrefix)
		require.NoError(t, err)

		return len(snapshotID) > 0, nil
	})
	require.NoError(t, err)

	return snapshotID[0].Name, err
}

// restoreV2Prov restores the cluster to the previous state after a snapshot is taken.
func restoreV2Prov(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, snapshotName string, terraformOptions *terraform.Options, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File, nestedRancherModuleDir string) {
	terratestConfig.SnapshotInput.CreateSnapshot = false
	terratestConfig.SnapshotInput.RestoreSnapshot = true
	terratestConfig.SnapshotInput.SnapshotName = snapshotName

	_, _, err := framework.ConfigTF(client, rancherConfig, terratestConfig, "", terraformConfig, newFile, rootBody, file, false, false, false, "", nestedRancherModuleDir)
	require.NoError(t, err)

	terraform.ApplyContext(t, context.Background(), terraformOptions)

	cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraformConfig.ResourcePrefix)
	require.NoError(t, err)

	err = provisioningActions.VerifyClusterReady(client, cluster)
	require.NoError(t, err)

	clusterObject, _, err := clusters.GetProvisioningClusterByName(client, terraformConfig.ResourcePrefix, namespace)
	require.NoError(t, err)

	logrus.Infof("Cluster version is restored to: %s", clusterObject.Spec.KubernetesVersion)
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
