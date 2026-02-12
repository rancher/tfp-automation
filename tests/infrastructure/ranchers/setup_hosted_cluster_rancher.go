package ranchers

import (
	"os"
	"testing"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	infraConfig "github.com/rancher/tests/validation/recurring/infrastructure/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/hosted"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/stretchr/testify/require"
)

// CreateHostedClusterRancher is a function that creates a Rancher setup, either via CLI or web application
func CreateHostedClusterRancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.HostedKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, err := hosted.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	if terraformConfig.Provider != providers.AKS && terraformConfig.Provider != providers.EKS {
		// For providers that do not have built-in DNS records, this will update the Rancher server URL.
		_, err = operations.ReplaceValue([]string{"rancher", "host"}, terraformConfig.Standalone.RancherHostname, cattleConfig)
		require.NoError(t, err)
	}

	rancherConfig, terraformConfig, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)
	infraConfig.WriteConfigToFile(os.Getenv(configEnvironmentKey), cattleConfig)

	testSession := session.NewSession()

	_, err = PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
	if err != nil {
		return err
	}

	return nil
}
