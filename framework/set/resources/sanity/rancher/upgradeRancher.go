package rancher

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	upgradeRancher = "upgrade_rancher"
)

// UpgradeRancher is a function that will upgrade the Rancher configurations in the main.tf file.
func UpgradeRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2ServerOnePublicIP string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.SanityKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/sanity/rancher/upgrade.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2ServerOnePublicIP, upgradeRancher)

	command := "bash -c '/tmp/upgrade.sh " + terraformConfig.Standalone.UpgradedRancherChartRepository + " " +
		terraformConfig.Standalone.UpgradedRancherRepo + " " + terraformConfig.Standalone.CertType + " " +
		terraformConfig.Standalone.RancherHostname + " " + terraformConfig.Standalone.UpgradedRancherTagVersion + " " +
		terraformConfig.Standalone.UpgradedRancherChartVersion + " " + terraformConfig.Standalone.UpgradedRancherImage

	if terraformConfig.Standalone.UpgradedRancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.UpgradedRancherAgentImage
	} else {
		command += " \"\""
	}

	if terraformConfig.Standalone.FeatureFlags != nil && terraformConfig.Standalone.FeatureFlags.UpgradedTurtles != "" {
		command += " " + terraformConfig.Standalone.FeatureFlags.UpgradedTurtles
	} else {
		command += " \"\""
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
