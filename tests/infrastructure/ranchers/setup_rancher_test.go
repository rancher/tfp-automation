package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/features"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	standaloneConfig *config.Standalone
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformOptions *terraform.Options
}

func (i *RancherTestSuite) TestCreateRancher() {
	i.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	i.rancherConfig, i.terraformConfig, i.terratestConfig, i.standaloneConfig = config.LoadTFPConfigs(i.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	_, err := sanity.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	if i.terraformConfig.Provider == providers.AWS {
		testSession := session.NewSession()
		i.session = testSession

		var client *rancher.Client
		var err error

		if i.standaloneConfig.FeatureFlags.MCM == "" {
			client, err = PostRancherSetup(i.T(), i.terraformOptions, i.rancherConfig, i.session, i.terraformConfig.Standalone.RancherHostname, keyPath, false, false)
			require.NoError(i.T(), err)
		}

		if i.standaloneConfig.FeatureFlags != nil {
			if i.standaloneConfig.FeatureFlags.Turtles != "" {
				toggleFeatureFlag(client, defaults.Turtles, i.standaloneConfig.FeatureFlags.Turtles)
			}
		}
	}
}

func toggleFeatureFlag(client *rancher.Client, feature string, toggledState string) {
	switch toggledState {
	case defaults.ToggledOff, "true":
		features.UpdateFeatureFlag(client, feature, false)
	case defaults.ToggledOn, "false":
		features.UpdateFeatureFlag(client, feature, true)
	}
}

func TestRancherTestSuite(t *testing.T) {
	suite.Run(t, new(RancherTestSuite))
}
