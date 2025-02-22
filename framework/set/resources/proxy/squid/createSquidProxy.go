package squid

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	installSquidProxy = "install_squid_proxy"
)

// CreateSquidProxy is a function that will set the squid proxy configurations in the main.tf file.
func CreateSquidProxy(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rke2BastionPublicDNS string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	scriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/proxy/squid/setup.sh")
	squidConf := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/proxy/squid/squid.conf")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}

	squidConfContent, err := os.ReadFile(squidConf)
	if err != nil {
		return nil, err
	}

	escapedSquidConfContent := strings.ReplaceAll(string(squidConfContent), "%", "%%")

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2BastionPublicDNS, installSquidProxy)

	command := "bash -c '/tmp/setup.sh " + terraformConfig.Standalone.OSUser + " " + rke2BastionPublicDNS + " " +
		terraformConfig.Standalone.BootstrapPassword + " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(scriptContent) + "' > /tmp/setup.sh"),
		cty.StringVal("printf '" + string(escapedSquidConfContent) + "' > /tmp/squid.conf"),
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
