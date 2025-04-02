package airgap

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	v2 "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke2k3s"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
)

const (
	airgapNodeOne       = "airgap_node1"
	airgapNodeTwo       = "airgap_node2"
	airgapNodeThree     = "airgap_node3"
	airgapWindowsNode   = "airgap_windows_node"
	bastion             = "bastion"
	copyScriptToBastion = "copy_script_to_bastion"
)

// SetAirgapRKE2K3s is a function that will set the airgap RKE2/K3s cluster configurations in the main.tf file.
func SetAirgapRKE2K3s(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	v2.SetRancher2ClusterV2(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, bastion)
	rootBody.AppendNewline()

	// Based on GH issue https://github.com/rancher/rancher/issues/45607, K3s clusters will only have one node.
	instances := []string{}
	if terraformConfig.Module == modules.AirgapRKE2 || terraformConfig.Module == modules.AirgapRKE2Windows {
		instances = []string{airgapNodeOne, airgapNodeTwo, airgapNodeThree}
	} else if terraformConfig.Module == modules.AirgapK3S {
		instances = []string{airgapNodeOne}
	}

	for _, instance := range instances {
		aws.CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	if strings.Contains(terraformConfig.Module, modules.AirgapRKE2Windows) {
		aws.CreateAirgappedWindowsAWSInstances(rootBody, terraformConfig, airgapWindowsNode)
		rootBody.AppendNewline()
	}

	provisionerBlockBody, err := nullresource.SetAirgapNullResource(rootBody, terraformConfig, copyScriptToBastion, nil)
	if err != nil {
		return nil, err
	}

	rootBody.AppendNewline()

	file, _ = locals.SetLocals(rootBody, terraformConfig, configMap, newFile, file, nil)

	rootBody.AppendNewline()

	err = copyScript(provisionerBlockBody)
	if err != nil {
		return nil, err
	}

	registrationCommands, nodePrivateIPs := GetRKE2K3sRegistrationCommands(terraformConfig)

	for _, instance := range instances {
		var dependsOn []string

		// Depending on the airgapped node, add the specific dependsOn expression.
		bastionScriptExpression := "[" + defaults.NullResource + `.copy_script_to_bastion` + "]"
		nodeOneExpression := "[" + defaults.NullResource + `.register_` + airgapNodeOne + "]"
		nodeTwoExpression := "[" + defaults.NullResource + `.register_` + airgapNodeTwo + "]"

		bastionPublicIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, bastion, defaults.PublicIp)

		if instance == airgapNodeOne {
			dependsOn = append(dependsOn, bastionScriptExpression)
		} else if instance == airgapNodeTwo {
			dependsOn = append(dependsOn, nodeOneExpression)
		} else if instance == airgapNodeThree {
			dependsOn = append(dependsOn, nodeTwoExpression)
		}

		provisionerBlockBody, err = nullresource.SetAirgapNullResource(rootBody, terraformConfig, "register_"+instance, dependsOn)
		if err != nil {
			return nil, err
		}

		err = registerPrivateNodes(provisionerBlockBody, terraformConfig, bastionPublicIP, nodePrivateIPs[instance], registrationCommands[instance])
		if err != nil {
			return nil, err
		}

		rootBody.AppendNewline()
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write airgap RKE2/K3s configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// GetRKE2K3sRegistrationCommands is a helper function that will return the registration commands for the airgap nodes.
func GetRKE2K3sRegistrationCommands(terraformConfig *config.TerraformConfig) (map[string]string, map[string]string) {
	commands := make(map[string]string)
	nodePrivateIPs := make(map[string]string)

	etcdRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.EtcdRoleFlag)
	controlPlaneRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.ControlPlaneRoleFlag)
	workerRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.WorkerRoleFlag)
	allRolesRegistrationCommand := fmt.Sprintf("${%s.%s_%s} %s", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureNodeCommand, defaults.AllFlags)
	windowsRegistrationCommand := fmt.Sprintf("${%s.%s_%s}", defaults.Local, terraformConfig.ResourcePrefix, defaults.InsecureWindowsNodeCommand)

	airgapNodeOnePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeOne, defaults.PrivateIp)
	airgapNodeTwoPrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeTwo, defaults.PrivateIp)
	airgapNodeThreePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeThree, defaults.PrivateIp)
	airgapWindowsNodePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapWindowsNode, defaults.PrivateIp)

	if terraformConfig.Module == modules.AirgapRKE2 || terraformConfig.Module == modules.AirgapRKE2Windows {
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
