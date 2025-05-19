package airgap

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke1"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
)

// // SetAirgapRKE1 is a function that will set the airgap RKE1 cluster configurations in the main.tf file.
func SetAirgapRKE1(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	rke1.SetRancher2Cluster(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, bastion+"_"+terraformConfig.ResourcePrefix)
	rootBody.AppendNewline()

	instances := []string{airgapNodeOne, airgapNodeTwo, airgapNodeThree}

	for _, instance := range instances {
		aws.CreateAirgappedAWSInstances(rootBody, terraformConfig, instance+"_"+terraformConfig.ResourcePrefix)
		rootBody.AppendNewline()
	}

	provisionerBlockBody, err := nullresource.SetAirgapNullResource(rootBody, terraformConfig, copyScriptToBastion+"_"+terraformConfig.ResourcePrefix, nil)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	err = copyScript(provisionerBlockBody, terraformConfig, terratestConfig)
	if err != nil {
		return nil, nil, err
	}

	registrationCommands, nodePrivateIPs := getRKE1RegistrationCommands(terraformConfig)

	for _, instance := range instances {
		var dependsOn []string

		// Depending on the airgapped node, add the specific dependsOn expression.
		bastionScriptExpression := "[" + defaults.NullResource + `.copy_script_to_bastion_` + terraformConfig.ResourcePrefix + "]"
		nodeOneExpression := "[" + defaults.NullResource + `.register_` + airgapNodeOne + "_" + terraformConfig.ResourcePrefix + "]"
		nodeTwoExpression := "[" + defaults.NullResource + `.register_` + airgapNodeTwo + "_" + terraformConfig.ResourcePrefix + "]"

		bastionPublicIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, bastion+"_"+terraformConfig.ResourcePrefix, defaults.PublicIp)

		if instance == airgapNodeOne {
			dependsOn = append(dependsOn, bastionScriptExpression)
		} else if instance == airgapNodeTwo {
			dependsOn = append(dependsOn, nodeOneExpression)
		} else if instance == airgapNodeThree {
			dependsOn = append(dependsOn, nodeTwoExpression)
		}

		provisionerBlockBody, err = nullresource.SetAirgapNullResource(rootBody, terraformConfig, "register_"+instance+"_"+terraformConfig.ResourcePrefix, dependsOn)
		if err != nil {
			return nil, nil, err
		}

		err = registerPrivateNodes(provisionerBlockBody, terraformConfig, bastionPublicIP, nodePrivateIPs[instance], registrationCommands[instance])
		if err != nil {
			return nil, nil, err
		}

		rootBody.AppendNewline()
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write airgap RKE1 configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}

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
