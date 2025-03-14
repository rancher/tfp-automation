package airgap

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/airgap/nullresource"
	"github.com/sirupsen/logrus"
)

// SetAirgapRKE2Windows is a function that will set the airgap RKE2 cluster configurations in the main.tf file.
func SetAirgapRKE2Windows(client *rancher.Client, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, configMap []map[string]any, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	provisionerBlockBody, err := nullresource.SetAirgapNullResource(rootBody, terraformConfig, "register_"+airgapWindowsNode, nil)
	rootBody.AppendNewline()

	bastionPublicIP := fmt.Sprintf("${%s.%s.%s}", defaults.AwsInstance, bastion, defaults.PublicIp)
	registrationCommands, nodePrivateIPs := GetRKE2K3sRegistrationCommands(terraformConfig)

	err = registerWindowsPrivateNodes(provisionerBlockBody, terraformConfig, bastionPublicIP, nodePrivateIPs[airgapWindowsNode], registrationCommands[airgapWindowsNode])
	if err != nil {
		return nil, err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write airgap Windows RKE2 configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
