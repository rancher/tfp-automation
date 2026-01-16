package rke1

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke1"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
)

const (
	airgapNodeOne       = "airgap_node1"
	airgapNodeTwo       = "airgap_node2"
	airgapNodeThree     = "airgap_node3"
	airgapWindowsNode   = "airgap_windows_node"
	bastion             = "bastion"
	copyScriptToBastion = "copy_script_to_bastion"
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

	err = airgap.CopyScript(provisionerBlockBody, terraformConfig, terratestConfig)
	if err != nil {
		return nil, nil, err
	}

	registrationCommands, nodePrivateIPs := getRKE1RegistrationCommands(terraformConfig)

	for _, instance := range instances {
		var dependsOn []string

		// Depending on the airgapped node, add the specific dependsOn expression.
		bastionScriptExpression := "[" + general.NullResource + `.copy_script_to_bastion_` + terraformConfig.ResourcePrefix + "]"
		nodeOneExpression := "[" + general.NullResource + `.register_` + airgapNodeOne + "_" + terraformConfig.ResourcePrefix + "]"
		nodeTwoExpression := "[" + general.NullResource + `.register_` + airgapNodeTwo + "_" + terraformConfig.ResourcePrefix + "]"

		switch instance {
		case airgapNodeOne:
			dependsOn = append(dependsOn, bastionScriptExpression)
		case airgapNodeTwo:
			dependsOn = append(dependsOn, nodeOneExpression)
		case airgapNodeThree:
			dependsOn = append(dependsOn, nodeTwoExpression)
		}

		provisionerBlockBody, err = nullresource.SetAirgapNullResource(rootBody, terraformConfig, "register_"+instance+"_"+terraformConfig.ResourcePrefix, dependsOn)
		if err != nil {
			return nil, nil, err
		}

		err = airgap.RegisterPrivateNodes(provisionerBlockBody, terraformConfig, nodePrivateIPs[instance], registrationCommands[instance])
		if err != nil {
			return nil, nil, err
		}

		rootBody.AppendNewline()
	}

	return newFile, file, nil
}
