package upgrade

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
	"github.com/sirupsen/logrus"
)

const (
	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating a Rancher server behind a proxy.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, terraformConfig *config.TerraformConfig,
	terratest *config.TerratestConfig, proxyNode, serverNode string) error {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	aws.CreateTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	aws.CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	file = sanity.OpenFile(file, keyPath)
	switch {
	case terraformConfig.Standalone.ProxyRancher:
		logrus.Infof("Upgrading Proxy Rancher...")
		_, err := rancher.UpgradeProxiedRancher(file, newFile, rootBody, terraformConfig, serverNode, proxyNode)
		if err != nil {
			return err
		}
	default:
		logrus.Errorf("Unsupported Rancher environment. Please check the configuration file.")
	}

	terraform.InitAndApply(t, terraformOptions)

	return nil
}
