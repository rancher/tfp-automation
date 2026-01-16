package ranchers

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/features"
	infraConfig "github.com/rancher/tests/validation/recurring/infrastructure/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework"
	featureDefaults "github.com/rancher/tfp-automation/framework/set/defaults/features"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/stretchr/testify/require"
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

	// For providers that do not have built-in DNS records, this will update the Rancher server URL.
	if terraformConfig.Provider != providers.AWS {
		_, err = operations.ReplaceValue([]string{"rancher", "host"}, terraformConfig.Standalone.RancherHostname, cattleConfig)
		require.NoError(t, err)

		rancherConfig, terraformConfig, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)
		infraConfig.WriteConfigToFile(os.Getenv(configEnvironmentKey), cattleConfig)
	}

	var client *rancher.Client
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
		toggleFeatureFlag(client, featureDefaults.Turtles, standaloneConfig.FeatureFlags.Turtles)
	}

	return nil
}

func toggleFeatureFlag(client *rancher.Client, feature string, toggledState string) {
	switch toggledState {
	case featureDefaults.ToggledOff, "true":
		features.UpdateFeatureFlag(client, feature, false)
	case featureDefaults.ToggledOn, "false":
		features.UpdateFeatureFlag(client, feature, true)
	}
}
