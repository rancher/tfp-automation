package rancher

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	upgradeRancher = "upgrade_airgap_rancher"
)

// UpgradeAirgapRancher is a function that will upgrade the Rancher configurations in the main.tf file.
func UpgradeAirgapRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	registryPublicDNS, bastionNode string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/airgap/rancher/upgrade.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, bastionNode, upgradeRancher)

	command := "bash -c '/tmp/upgrade.sh " + terraformConfig.Standalone.UpgradedRancherChartRepository + " " +
		terraformConfig.Standalone.UpgradedRancherRepo + " " + terraformConfig.Standalone.RancherHostname + " " +
		terraformConfig.Standalone.AirgapInternalFQDN + " " + terraformConfig.Standalone.UpgradedRancherTagVersion + " " +
		terraformConfig.Standalone.BootstrapPassword + " " + terraformConfig.Standalone.UpgradedRancherImage + " " + registryPublicDNS

	if terraformConfig.Standalone.UpgradedRancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.UpgradedRancherAgentImage
	}

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(scriptContent) + "' > /tmp/upgrade.sh"),
		cty.StringVal("chmod +x /tmp/upgrade.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
