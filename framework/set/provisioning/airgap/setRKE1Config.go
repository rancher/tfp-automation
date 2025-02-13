package airgap

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/locals"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke1"
	airgap "github.com/rancher/tfp-automation/framework/set/resources/airgap/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
	"github.com/sirupsen/logrus"
)

// // SetAirgapRKE1 is a function that will set the airgap RKE1 cluster configurations in the main.tf file.
func SetAirgapRKE1(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	configMap []map[string]any, clusterName string, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	rke1.SetRancher2Cluster(rootBody, terraformConfig, clusterName)
	rootBody.AppendNewline()

	aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, bastion)
	rootBody.AppendNewline()

	instances := []string{airgapNodeOne, airgapNodeTwo, airgapNodeThree}

	for _, instance := range instances {
		airgap.CreateAirgappedAWSInstances(rootBody, terraformConfig, instance)
		rootBody.AppendNewline()
	}

	provisionerBlockBody, err := nullresource.SetAirgapNullResource(rootBody, terraformConfig, copyScriptToBastion, nil)
	if err != nil {
		return nil, err
	}

	rootBody.AppendNewline()

	file, _ = locals.SetLocals(rootBody, terraformConfig, configMap, clusterName, newFile, file, nil)

	rootBody.AppendNewline()

	err = copyScript(provisionerBlockBody)
	if err != nil {
		return nil, err
	}

	registrationCommands, nodePrivateIPs := getRKE1RegistrationCommands(clusterName)

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

// getRKE1RegistrationCommands is a helper function that will return the registration commands for the airgap nodes.
func getRKE1RegistrationCommands(clusterName string) (map[string]string, map[string]string) {
	commands := make(map[string]string)
	nodePrivateIPs := make(map[string]string)

	regCommand := fmt.Sprintf("${%s.%s.%s[0].%s}", defaults.Cluster, clusterName, defaults.ClusterRegistrationToken, defaults.NodeCommand)

	etcdRegistrationCommand := fmt.Sprintf(regCommand+" %s", defaults.EtcdRoleFlag)
	controlPlaneRegistrationCommand := fmt.Sprintf(regCommand+" %s", defaults.ControlPlaneRoleFlag)
	workerRegistrationCommand := fmt.Sprintf(regCommand+" %s", defaults.WorkerRoleFlag)

	airgapNodeOnePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeOne, defaults.PrivateIp)
	airgapNodeTwoPrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeTwo, defaults.PrivateIp)
	airgapNodeThreePrivateIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, airgapNodeThree, defaults.PrivateIp)

	commands[airgapNodeOne] = etcdRegistrationCommand
	commands[airgapNodeTwo] = controlPlaneRegistrationCommand
	commands[airgapNodeThree] = workerRegistrationCommand

	nodePrivateIPs[airgapNodeOne] = airgapNodeOnePrivateIP
	nodePrivateIPs[airgapNodeTwo] = airgapNodeTwoPrivateIP
	nodePrivateIPs[airgapNodeThree] = airgapNodeThreePrivateIP

	return commands, nodePrivateIPs
}
