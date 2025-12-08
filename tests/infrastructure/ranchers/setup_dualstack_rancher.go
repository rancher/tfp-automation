package ranchers

import (
	"os"
	"testing"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/dualstack"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
)

// CreateDualStackRancher is a function that creates a Rancher setup, either via CLI or web application
func CreateDualStackRancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.DualStackKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, err := dualstack.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	testSession := session.NewSession()

	_, err = PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
	if err != nil {
		return err
	}

	return nil
}
