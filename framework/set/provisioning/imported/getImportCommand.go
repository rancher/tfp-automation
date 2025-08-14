package imported

import (
	"fmt"

	"github.com/rancher/tfp-automation/framework/set/defaults"
)

const (
	serverOne = "server1"
)

// GetImportCommand is a helper function that will return the import command for the cluster
func GetImportCommand(clusterName string) map[string]string {
	command := make(map[string]string)
	importCommand := fmt.Sprintf("${%s.%s.%s[0].%s}", defaults.Cluster, clusterName, defaults.ClusterRegistrationToken, defaults.InsecureCommand)

	serverOneName := clusterName + `_` + serverOne
	command[serverOneName] = importCommand

	return command
}
