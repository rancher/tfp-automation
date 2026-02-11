package rancher

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	installRancher = "install_rancher"
)

// CreateRancher is a function that will set the Rancher configurations in the main.tf file.
func CreateRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, bastionPublicIP string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.SanityKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/hosted/rancher/setup.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, bastionPublicIP, installRancher)

	command := "/tmp/setup.sh " + terraformConfig.ResourcePrefix + " " + terraformConfig.Provider + " " +
		terraformConfig.Standalone.RancherChartRepository + " " + terraformConfig.Standalone.Repo + " " +
		terraformConfig.Standalone.CertManagerVersion + " " + terraformConfig.Standalone.RancherHostname + " " +
		terraformConfig.Standalone.RancherTagVersion + " " + terraformConfig.Standalone.ChartVersion + " " +
		terraformConfig.Standalone.BootstrapPassword + " " + terraformConfig.Standalone.RancherImage

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	} else {
		command += " \"\""
	}

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("cat <<'EOF' > /tmp/setup.sh\n" + string(scriptContent) + "\nEOF"),
		cty.StringVal("chmod +x /tmp/setup.sh"),
		cty.StringVal(command),
	}))

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}
