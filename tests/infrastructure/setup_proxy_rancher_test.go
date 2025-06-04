package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	resources "github.com/rancher/tfp-automation/framework/set/resources/proxy"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProxyRancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *ProxyRancherTestSuite) TestCreateProxyRancher() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	_, _, err := resources.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	logrus.Infof("Rancher server URL: %s", i.terraformConfig.Standalone.RancherHostname)
	logrus.Infof("Booststrap password: %s", i.terraformConfig.Standalone.BootstrapPassword)
	logrus.Infof("Proxy Address: %s:3228", i.terraformConfig.Proxy.ProxyBastion)

	testSession := session.NewSession()
	i.session = testSession

	_, err = PostRancherSetup(i.T(), i.session, i.terraformConfig.Standalone.RancherHostname, false, false)
	require.NoError(i.T(), err)
}

func TestProxyRancherTestSuite(t *testing.T) {
	suite.Run(t, new(ProxyRancherTestSuite))
}
