package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *RancherTestSuite) TestCreateRancher() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	_, err := resources.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	if i.terraformConfig.Provider != providers.Linode && i.terraformConfig.Provider != providers.Vsphere {
		logrus.Infof("Rancher server URL: %s", i.terraformConfig.Standalone.RancherHostname)
		logrus.Infof("Booststrap password: %s", i.terraformConfig.Standalone.BootstrapPassword)

		testSession := session.NewSession()
		i.session = testSession

		_, err = PostRancherSetup(i.T(), i.session, i.terraformConfig.Standalone.RancherHostname, false, false)
		require.NoError(i.T(), err)
	}
}

func TestRancherTestSuite(t *testing.T) {
	suite.Run(t, new(RancherTestSuite))
}
