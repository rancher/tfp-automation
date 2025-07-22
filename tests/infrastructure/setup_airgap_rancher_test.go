package infrastructure

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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AirgapRancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *AirgapRancherTestSuite) TestCreateAirgapRancher() {
	i.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	i.rancherConfig, i.terraformConfig, i.terratestConfig = config.LoadTFPConfigs(i.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	registry, _, err := resources.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	logrus.Infof("Rancher server URL: %s", i.terraformConfig.Standalone.RancherHostname)
	logrus.Infof("Booststrap password: %s", i.terraformConfig.Standalone.BootstrapPassword)
	logrus.Infof("Private registry: %s", registry)

	testSession := session.NewSession()
	i.session = testSession

	_, err = PostRancherSetup(i.T(), i.rancherConfig, i.session, i.terraformConfig.Standalone.AirgapInternalFQDN, false, true)
	require.NoError(i.T(), err)
}

func TestAirgapRancherTestSuite(t *testing.T) {
	suite.Run(t, new(AirgapRancherTestSuite))
}
