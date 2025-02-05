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
	installRancher = "install_proxy_rancher"
)

// CreateProxiedRancher is a function that will set the Rancher configurations in the main.tf file.
func CreateProxiedRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rke2ServerOnePublicDNS, rke2BastionPublicDNS string) (*os.File, error) {

	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher/setup.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2ServerOnePublicDNS, installRancher)

	command := "bash -c '/tmp/setup.sh " + terraformConfig.Standalone.RancherChartRepository + " " +
		terraformConfig.Standalone.Repo + " " + terraformConfig.Standalone.CertManagerVersion + " " +
		terraformConfig.Standalone.RancherHostname + " " + terraformConfig.Standalone.RancherTagVersion + " " +
		terraformConfig.Standalone.BootstrapPassword + " " + terraformConfig.Standalone.RancherImage + " " +
		rke2BastionPublicDNS

	if terraformConfig.Standalone.StagingRancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.StagingRancherAgentImage
	}

	command += "'"

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
