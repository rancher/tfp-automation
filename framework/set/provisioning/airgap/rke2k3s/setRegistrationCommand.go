package rke2k3s

import (
	"fmt"
	"strings"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
)

// getRKE2K3sRegistrationCommands is a helper function that will return the registration commands for the airgap nodes.
func getRKE2K3sRegistrationCommands(terraformConfig *config.TerraformConfig) (map[string]string, map[string]string) {
	commands := make(map[string]string)
	nodePrivateIPs := make(map[string]string)

	etcdRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", general.Local, terraformConfig.ResourcePrefix, clusters.InsecureNodeCommand, clusters.EtcdRoleFlag)
	controlPlaneRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", general.Local, terraformConfig.ResourcePrefix, clusters.InsecureNodeCommand, clusters.ControlPlaneRoleFlag)
	workerRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", general.Local, terraformConfig.ResourcePrefix, clusters.InsecureNodeCommand, clusters.WorkerRoleFlag)
	allRolesRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", general.Local, terraformConfig.ResourcePrefix, clusters.InsecureNodeCommand, clusters.AllFlags)
	windowsRegistrationCommand := fmt.Sprintf("${%s.%s_%s}", general.Local, terraformConfig.ResourcePrefix, clusters.InsecureWindowsNodeCommand)
	airgapNodeOnePrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapNodeOne+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)
	airgapNodeTwoPrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapNodeTwo+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)
	airgapNodeThreePrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapNodeThree+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)
	airgapWindowsNodePrivateIP := fmt.Sprintf("${%s.%s.%s}", aws.AwsInstance, airgapWindowsNode+"_"+terraformConfig.ResourcePrefix, general.PrivateIp)

	if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		commands[airgapNodeOne] = etcdRegistrationCommand
		commands[airgapNodeTwo] = controlPlaneRegistrationCommand
		commands[airgapNodeThree] = workerRegistrationCommand
		commands[airgapWindowsNode] = windowsRegistrationCommand

		nodePrivateIPs[airgapNodeOne] = airgapNodeOnePrivateIP
		nodePrivateIPs[airgapNodeTwo] = airgapNodeTwoPrivateIP
		nodePrivateIPs[airgapNodeThree] = airgapNodeThreePrivateIP
		nodePrivateIPs[airgapWindowsNode] = airgapWindowsNodePrivateIP
	} else if terraformConfig.Module == modules.AirgapK3S {
		commands[airgapNodeOne] = allRolesRegistrationCommand
		nodePrivateIPs[airgapNodeOne] = airgapNodeOnePrivateIP
	}

	return commands, nodePrivateIPs
}
