package upgrade

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	airgap "github.com/rancher/tfp-automation/framework/set/resources/airgap/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	proxy "github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher"
	registry "github.com/rancher/tfp-automation/framework/set/resources/registries/createRegistry"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	sanityRancher "github.com/rancher/tfp-automation/framework/set/resources/sanity/rancher"
	"github.com/sirupsen/logrus"
)

const (
	nonAuthRegistry = "non_auth_registry"
	terraformConst  = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating a Rancher server behind a proxy.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, rancherConfig *rancher.Config,
	terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, serverNode, proxyNode, bastionNode,
	registryNode string) error {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	aws.CreateAWSTerraformProviderBlock(tfBlockBody)
	rootBody.AppendNewline()

	aws.CreateAWSProviderBlock(rootBody, terraformConfig)
	rootBody.AppendNewline()

	file = sanity.OpenFile(file, keyPath)
	switch {
	case terraformConfig.Standalone.UpgradeAirgapRancher:
		logrus.Infof("Updating private registry...")
		_, err := registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, registryNode, nonAuthRegistry)
		if err != nil {
			return err
		}

		_, err = terraform.InitAndApplyE(t, terraformOptions)
		if err != nil && *rancherConfig.Cleanup {
			logrus.Infof("Error while updating private registry. Cleaning up...")
			cleanup.Cleanup(t, terraformOptions, keyPath)
			return err
		}

		file = sanity.OpenFile(file, keyPath)
		logrus.Infof("Upgrading Airgap Rancher...")
		file, err = airgap.UpgradeAirgapRancher(file, newFile, rootBody, terraformConfig, terratestConfig, registryNode, bastionNode)
		if err != nil {
			return err
		}

		_, err = terraform.InitAndApplyE(t, terraformOptions)
		if err != nil && *rancherConfig.Cleanup {
			logrus.Infof("Error while upgrading Airgap Rancher. Cleaning up...")
			cleanup.Cleanup(t, terraformOptions, keyPath)
			return err
		}
	case terraformConfig.Standalone.UpgradeProxyRancher:
		logrus.Infof("Upgrading Proxy Rancher...")
		_, err := proxy.UpgradeProxiedRancher(file, newFile, rootBody, terraformConfig, terratestConfig, serverNode, proxyNode)
		if err != nil {
			return err
		}

		_, err = terraform.InitAndApplyE(t, terraformOptions)
		if err != nil && *rancherConfig.Cleanup {
			logrus.Infof("Error while upgrading Proxy Rancher. Cleaning up...")
			cleanup.Cleanup(t, terraformOptions, keyPath)
			return err
		}
	case terraformConfig.Standalone.UpgradeRancher, terraformConfig.Standalone.UpgradeDualStackRancher, terraformConfig.Standalone.UpgradeIPv6Rancher:
		logrus.Infof("Upgrading Rancher...")
		_, err := sanityRancher.UpgradeRancher(file, newFile, rootBody, terraformConfig, terratestConfig, serverNode)
		if err != nil {
			return err
		}

		_, err = terraform.InitAndApplyE(t, terraformOptions)
		if err != nil && *rancherConfig.Cleanup {
			logrus.Infof("Error while upgrading Rancher. Cleaning up...")
			cleanup.Cleanup(t, terraformOptions, keyPath)
			return err
		}
	default:
		logrus.Errorf("Unsupported Rancher environment. Please check the configuration file.")
	}

	return nil
}
