package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	resources "github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AirgapRancherTestSuite struct {
	suite.Suite
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *AirgapRancherTestSuite) TestCreateAirgapRancher() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)

	keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	registry, _, err := resources.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	logrus.Infof("Rancher server URL: %s", i.terraformConfig.Standalone.RancherHostname)
	logrus.Infof("Booststrap password: %s", i.terraformConfig.Standalone.BootstrapPassword)
	logrus.Infof("Private registry: %s", registry)
}

func TestAirgapRancherTestSuite(t *testing.T) {
	suite.Run(t, new(AirgapRancherTestSuite))
}
