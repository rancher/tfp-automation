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
	installRancher = "install_rancher"
)

// CreateRancher is a function that will set the Rancher configurations in the main.tf file.
func CreateRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2ServerOnePublicIP, nodeBalancerHostname string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/sanity/rancher/setup.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2ServerOnePublicIP, installRancher)

	if nodeBalancerHostname != "" {
		terraformConfig.Standalone.RancherHostname = nodeBalancerHostname
	}

	command := "bash -c '/tmp/setup.sh " + terraformConfig.Standalone.RancherChartRepository + " " +
		terraformConfig.Standalone.Repo + " " + terraformConfig.Standalone.CertManagerVersion + " " +
		terraformConfig.Standalone.RancherHostname + " " + terraformConfig.Standalone.RancherTagVersion + " " +
		terraformConfig.Standalone.BootstrapPassword + " " + terraformConfig.Standalone.RancherImage

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	}

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(scriptContent) + "' > /tmp/setup.sh"),
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
