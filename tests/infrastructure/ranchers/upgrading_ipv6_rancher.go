package ranchers

import (
	"os"
	"testing"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	resources "github.com/rancher/tfp-automation/framework/set/resources/ipv6"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	"github.com/stretchr/testify/require"
)

// UpgradingIPv6Rancher is a function that creates and upgrades a IPv6 Rancher setup, either via CLI or web application
func UpgradingIPv6Rancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.IPv6KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	serverNodeOne, err := resources.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	testSession := session.NewSession()

	terraformConfig.Standalone.UpgradeIPv6Rancher = true

	_, upgradeKeyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, upgradeKeyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, upgradeKeyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.IPv6KeyPath)

	_, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	return nil
}
