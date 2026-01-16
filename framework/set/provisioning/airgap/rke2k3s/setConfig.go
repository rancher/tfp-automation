package rke2k3s

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
	v2 "github.com/rancher/tfp-automation/framework/set/provisioning/custom/rke2k3s"
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

// SetAirgapRKE2K3s is a function that will set the airgap RKE2/K3s cluster configurations in the main.tf file.
func SetAirgapRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	v2.SetRancher2ClusterV2(rootBody, terraformConfig, terratestConfig)
	rootBody.AppendNewline()

	aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, bastion+"_"+terraformConfig.ResourcePrefix)
	rootBody.AppendNewline()

	// Based on GH issue https://github.com/rancher/rancher/issues/45607, K3s clusters will only have one node.
	instances := []string{}
	if strings.Contains(terraformConfig.Module, clustertypes.RKE2) {
		instances = []string{airgapNodeOne, airgapNodeTwo, airgapNodeThree}
	} else if terraformConfig.Module == modules.AirgapK3S {
		instances = []string{airgapNodeOne}
	}

	for _, instance := range instances {
		aws.CreateAirgappedAWSInstances(rootBody, terraformConfig, instance+"_"+terraformConfig.ResourcePrefix)
		rootBody.AppendNewline()
	}

	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) {
		aws.CreateAirgappedWindowsAWSInstances(rootBody, terraformConfig, airgapWindowsNode+"_"+terraformConfig.ResourcePrefix)
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

	registrationCommands, nodePrivateIPs := getRKE2K3sRegistrationCommands(terraformConfig)

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
