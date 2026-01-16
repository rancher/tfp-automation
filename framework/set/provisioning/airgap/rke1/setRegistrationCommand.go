package rke1

import (
	"fmt"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
)

// getRKE1RegistrationCommands is a helper function that will return the registration commands for the airgap nodes.
func getRKE1RegistrationCommands(terraformConfig *config.TerraformConfig) (map[string]string, map[string]string) {
	commands := make(map[string]string)
	nodePrivateIPs := make(map[string]string)

	regCommand := fmt.Sprintf("${%s.%s.%s[0].%s}", rancher2.Cluster, terraformConfig.ResourcePrefix, clusters.ClusterRegistrationToken, clusters.NodeCommand)

	etcdRegistrationCommand := fmt.Sprintf(regCommand+" %s", clusters.EtcdRoleFlag)
	controlPlaneRegistrationCommand := fmt.Sprintf(regCommand+" %s", clusters.ControlPlaneRoleFlag)
	workerRegistrationCommand := fmt.Sprintf(regCommand+" %s", clusters.WorkerRoleFlag)
	airgapNodeOnePrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapNodeOne+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)
	airgapNodeTwoPrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapNodeTwo+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)
	airgapNodeThreePrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapNodeThree+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)

	commands[airgapNodeOne] = etcdRegistrationCommand
	commands[airgapNodeTwo] = controlPlaneRegistrationCommand
	commands[airgapNodeThree] = workerRegistrationCommand

	nodePrivateIPs[airgapNodeOne] = airgapNodeOnePrivateIP
	nodePrivateIPs[airgapNodeTwo] = airgapNodeTwoPrivateIP
	nodePrivateIPs[airgapNodeThree] = airgapNodeThreePrivateIP

	return commands, nodePrivateIPs
}
