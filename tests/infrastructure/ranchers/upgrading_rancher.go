package ranchers

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	"github.com/stretchr/testify/require"
)

// UpgradingRancher is a function that creates and upgrades a Rancher setup, either via CLI or web application
func UpgradingRancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	serverNodeOne, err := resources.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	testSession := session.NewSession()

	var client *rancher.Client

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.MCM == "" {
		client, err = PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
		require.NoError(t, err)
	} else if standaloneConfig.FeatureFlags == nil {
		client, err = PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
		require.NoError(t, err)
	}

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
		toggleFeatureFlag(client, defaults.Turtles, standaloneConfig.FeatureFlags.Turtles)
	}

	terraformConfig.Standalone.UpgradeRancher = true

	_, upgradeKeyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, upgradeKeyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, upgradeKeyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.SanityKeyPath)

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.UpgradedMCM == "" {
		client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, true)
		require.NoError(t, err)
	} else if standaloneConfig.FeatureFlags == nil {
		client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, true)
		require.NoError(t, err)
	}

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.UpgradedTurtles != "" {
		toggleFeatureFlag(client, defaults.Turtles, standaloneConfig.FeatureFlags.UpgradedTurtles)
	}

	return nil
}
