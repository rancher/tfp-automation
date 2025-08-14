package rke2k3s

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
)

// SetAirgapRKE2Windows is a function that will set the airgap RKE2 cluster configurations in the main.tf file.
func SetAirgapRKE2Windows(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	provisionerBlockBody, err := nullresource.SetAirgapNullResource(rootBody, terraformConfig, "register_"+airgapWindowsNode+"_"+terraformConfig.ResourcePrefix, nil)
	rootBody.AppendNewline()

	registrationCommands, nodePrivateIPs := getRKE2K3sRegistrationCommands(terraformConfig)

	err = airgap.RegisterWindowsPrivateNodes(provisionerBlockBody, terraformConfig, nodePrivateIPs[airgapWindowsNode], registrationCommands[airgapWindowsNode])
	if err != nil {
		return nil, nil, err
	}

	return newFile, file, nil
}
