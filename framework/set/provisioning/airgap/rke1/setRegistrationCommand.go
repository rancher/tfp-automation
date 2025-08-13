package rke1

import (
	"fmt"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
)

// getRKE1RegistrationCommands is a helper function that will return the registration commands for the airgap nodes.
func getRKE1RegistrationCommands(terraformConfig *config.TerraformConfig) (map[string]string, map[string]string) {
	commands := make(map[string]string)
	nodePrivateIPs := make(map[string]string)

	regCommand := fmt.Sprintf("${%s.%s.%s[0].%s}", defaults.Cluster, terraformConfig.ResourcePrefix, defaults.ClusterRegistrationToken, defaults.NodeCommand)

	etcdRegistrationCommand := fmt.Sprintf(regCommand+" %s", defaults.EtcdRoleFlag)
	controlPlaneRegistrationCommand := fmt.Sprintf(regCommand+" %s", defaults.ControlPlaneRoleFlag)
	workerRegistrationCommand := fmt.Sprintf(regCommand+" %s", defaults.WorkerRoleFlag)

	airgapNodeOnePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeOne+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)
	airgapNodeTwoPrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeTwo+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)
	airgapNodeThreePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeThree+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)

	commands[airgapNodeOne] = etcdRegistrationCommand
	commands[airgapNodeTwo] = controlPlaneRegistrationCommand
	commands[airgapNodeThree] = workerRegistrationCommand

	nodePrivateIPs[airgapNodeOne] = airgapNodeOnePrivateIP
	nodePrivateIPs[airgapNodeTwo] = airgapNodeTwoPrivateIP
	nodePrivateIPs[airgapNodeThree] = airgapNodeThreePrivateIP

	return commands, nodePrivateIPs
}
