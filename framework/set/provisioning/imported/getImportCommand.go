package imported

import (
	"fmt"

	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
)

const (
	serverOne = "server1"
)

// GetImportCommand is a helper function that will return the import command for the cluster
func GetImportCommand(clusterName string) map[string]string {
	command := make(map[string]string)
	importCommand := fmt.Sprintf("${%s.%s.%s[0].%s}", rancher2.Cluster, clusterName, clusters.ClusterRegistrationToken, clusters.InsecureCommand)

	serverOneName := clusterName + `_` + serverOne
	command[serverOneName] = importCommand

	return command
}
