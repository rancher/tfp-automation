package rancher

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	installRancher = "install_rancher"
)

// CreateRancher is a function that will set the Rancher configurations in the main.tf file.
func CreateRancher(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2ServerOnePublicDNS string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/sanity/rancher/setup.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2ServerOnePublicDNS, installRancher)

	command := "bash -c '/tmp/setup.sh " + terraformConfig.Standalone.RKE2User + " " + terraformConfig.Standalone.RKE2Group + " " +
		terraformConfig.Standalone.RancherRepo + " " + terraformConfig.Standalone.RancherChartRepository + " " +
		terraformConfig.Standalone.Type + " " + terraformConfig.Standalone.CertManagerVersion + " " +
		terraformConfig.Standalone.RancherHostname + " " + terraformConfig.Standalone.RancherTagVersion + " " +
		terraformConfig.Standalone.BootstrapPassword

	if terraformConfig.Standalone.StagingRancherImage != "" {
		command += " " + terraformConfig.Standalone.StagingRancherImage + " " + terraformConfig.Standalone.StagingRancherAgentImage
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
