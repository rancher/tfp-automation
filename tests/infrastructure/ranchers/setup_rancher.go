package ranchers

import (
	"os"
	"testing"

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
)

// CreateRancher is a function that creates a Rancher setup, either via CLI or web application
func CreateRancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, err := sanity.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	if terraformConfig.Provider == providers.AWS {
		var client *rancher.Client
		var err error

		testSession := session.NewSession()

		if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.MCM == "" {
			client, err = PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
			if err != nil {
				return err
			}
		} else if standaloneConfig.FeatureFlags == nil {
			client, err = PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
			if err != nil {
				return err
			}
		}

		if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
			toggleFeatureFlag(client, defaults.Turtles, standaloneConfig.FeatureFlags.Turtles)
		}
	}

	return nil
}

func toggleFeatureFlag(client *rancher.Client, feature string, toggledState string) {
	switch toggledState {
	case defaults.ToggledOff, "true":
		features.UpdateFeatureFlag(client, feature, false)
	case defaults.ToggledOn, "false":
		features.UpdateFeatureFlag(client, feature, true)
	}
}
