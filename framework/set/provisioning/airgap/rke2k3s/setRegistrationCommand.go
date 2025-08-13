package rke2k3s

import (
	"fmt"
	"strings"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
)

// getRKE2K3sRegistrationCommands is a helper function that will return the registration commands for the airgap nodes.
func getRKE2K3sRegistrationCommands(terraformConfig *config.TerraformConfig) (map[string]string, map[string]string) {
	commands := make(map[string]string)
	nodePrivateIPs := make(map[string]string)

	etcdRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.EtcdRoleFlag)
	controlPlaneRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.ControlPlaneRoleFlag)
	workerRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.WorkerRoleFlag)
	allRolesRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.AllFlags)
	windowsRegistrationCommand := fmt.Sprintf("${%s.%s_%s}", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureWindowsNodeCommand)

	airgapNodeOnePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeOne+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)
	airgapNodeTwoPrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeTwo+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)
	airgapNodeThreePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeThree+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)
	airgapWindowsNodePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapWindowsNode+"_"+terraformConfig.ResourcePrefix, defaults.PrivateIp)

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
