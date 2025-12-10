package ranchers

import (
	"os"
	"testing"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/features"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	"github.com/stretchr/testify/require"
)

// CreateAirgapRancher is a function that creates an airgap Rancher setup, either via CLI or web application
func CreateAirgapRancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, bastion, err := airgap.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	sshKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	require.NoError(t, err)

	err = ssh.StartBastionSSHTunnel(bastion, terraformConfig.Standalone.OSUser, sshKey, "8443", standaloneConfig.RancherHostname, "443")
	require.NoError(t, err)

	testSession := session.NewSession()

	client, err := PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(t, err)

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
		switch standaloneConfig.FeatureFlags.Turtles {
		case defaults.ToggledOff:
			features.UpdateFeatureFlag(client, defaults.Turtles, false)
		case defaults.ToggledOn:
			features.UpdateFeatureFlag(client, defaults.Turtles, true)
		}
	}

	return nil
}
