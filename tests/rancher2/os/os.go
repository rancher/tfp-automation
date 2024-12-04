package os

import (
	"github.com/rancher/rancher/tests/v2/actions/workloads/cronjob"
	"github.com/rancher/rancher/tests/v2/actions/workloads/daemonset"
	"github.com/rancher/rancher/tests/v2/actions/workloads/deployment"
	"github.com/rancher/rancher/tests/v2/actions/workloads/statefulset"
	"github.com/rancher/shepherd/clients/rancher"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/pkg/nodes"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SSHCluster struct {
	id    string
	nodes []*nodes.Node
}

func workloadTests(p *suite.Suite, client *rancher.Client, clusterIDs []string) {
	workloadTests := []struct {
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

	for _, workloadTest := range workloadTests {
		p.Run(workloadTest.name, func() {
			for _, clusterID := range clusterIDs {
				clusterName, err := clusterExtensions.GetClusterNameByID(client, clusterID)
				require.NoError(p.T(), err)

				logrus.Infof("Running %s on cluster %s", workloadTest.name, clusterName)
				retries := 3
				for i := 0; i+1 < retries; i++ {
					err := workloadTest.validationFunc(client, clusterID)
					if err != nil {
						logrus.Info(err)
						logrus.Infof("Retry %v / %v", i+1, retries)
						continue
					}

					break
				}
			}
		})
	}
}

/*
// NodeRebootTest reboots all nodes in the provided clusters, one per cluster at a time.
func NodeRebootTest(client *rancher.Client, clusterIDs []string, cattleConfig map[string]any) error {
	var clusters []*v1.SteveAPIObject
	var err error

	for _, clusterID := range clusterIDs {
		cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(clusterID)
		if err != nil {
			return err
		}

		clusters = append(clusters, cluster)
	}

	var SSHClusters []SSHCluster
	maxNodeNum := 0
	for _, cluster := range clusters {
		sshUser, err := sshkeys.GetSSHUser(client, cluster)
		if err != nil {
			return err
		}

		steveClient, err := steve.GetClusterClient(client, cluster.ID)
		if err != nil {
			return err
		}

		nodesSteveObjList, err := steveClient.SteveType(stevetypes.Node).List(nil)
		if err != nil {
			return err
		}

		var sshNodes []*nodes.Node
		for _, node := range nodesSteveObjList.Data {
			clusterNode, err := sshkeys.GetSSHNodeFromMachine(client, sshUser, &node)
			if err != nil {
				return err
			}

			sshNodes = append(sshNodes, clusterNode)
		}

		if len(sshNodes) > maxNodeNum {
			maxNodeNum = len(sshNodes)
		}

		SSHClusters = append(SSHClusters, SSHCluster{id: cluster.ID, nodes: sshNodes})
	}

	for i := range maxNodeNum {
		for _, cluster := range SSHClusters {
			if i > len(cluster.nodes) {
				continue
			}

			err := ec2.RebootNode(client, *cluster.nodes[i], cluster.id)
			if err != nil {
				return err
			}
		}

		for _, cluster := range clusters {
			err := steve.WaitForResourceState(client.Steve, cluster, stevestates.Active, time.Second, defaults.FifteenMinuteTimeout)
			if err != nil {
				return err
			}
		}
	}

	return err
}*/
