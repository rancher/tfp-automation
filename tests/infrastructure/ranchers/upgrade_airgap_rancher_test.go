package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	resources "github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UpgradeAirgapRancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	standaloneConfig *config.Standalone
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *UpgradeAirgapRancherTestSuite) TestUpgradeAirgapRancher() {
	i.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	i.rancherConfig, i.terraformConfig, i.terratestConfig, i.standaloneConfig = config.LoadTFPConfigs(i.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	registry, bastion, err := resources.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	testSession := session.NewSession()
	i.session = testSession

	_, err = PostRancherSetup(i.T(), i.terraformOptions, i.rancherConfig, i.session, i.terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(i.T(), err)

	i.terraformConfig.Standalone.UpgradeAirgapRancher = true

	_, upgradeKeyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, upgradeKeyPath)

	err = upgrade.CreateMainTF(i.T(), upgradeTerraformOptions, upgradeKeyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig, "", "", bastion, registry)
	require.NoError(i.T(), err)

	standaloneTerraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keypath.AirgapKeyPath)
	_, err = PostRancherSetup(i.T(), standaloneTerraformOptions, i.rancherConfig, i.session, i.terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(i.T(), err)
}

func TestUpgradeAirgapRancherTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeAirgapRancherTestSuite))
}
