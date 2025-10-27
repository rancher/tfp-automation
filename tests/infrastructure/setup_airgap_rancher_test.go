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
	"github.com/rancher/tests/actions/featureflags"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	resources "github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AirgapRancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	standaloneConfig *config.Standalone
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *AirgapRancherTestSuite) TestCreateAirgapRancher() {
	i.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	i.rancherConfig, i.terraformConfig, i.terratestConfig, i.standaloneConfig = config.LoadTFPConfigs(i.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	_, _, err := resources.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	testSession := session.NewSession()
	i.session = testSession

	client, err := PostRancherSetup(i.T(), i.terraformOptions, i.rancherConfig, i.session, i.terraformConfig.Standalone.AirgapInternalFQDN, keyPath, true, false)
	require.NoError(i.T(), err)

	if i.standaloneConfig.FeatureFlags != nil && i.standaloneConfig.FeatureFlags.Turtles != "" {
		switch i.standaloneConfig.FeatureFlags.Turtles {
		case defaults.ToggledOff:
			featureflags.UpdateFeatureFlag(client, defaults.Turtles, false)
		case defaults.ToggledOn:
			featureflags.UpdateFeatureFlag(client, defaults.Turtles, true)
		}
	}
}

func TestAirgapRancherTestSuite(t *testing.T) {
	suite.Run(t, new(AirgapRancherTestSuite))
}
