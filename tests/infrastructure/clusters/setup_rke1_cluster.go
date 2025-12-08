package clusters

import (
	"os"
	"testing"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	rke "github.com/rancher/tfp-automation/framework/set/resources/rke"
	"github.com/stretchr/testify/require"
)

// CreateRKE1Cluster is a function that creates a RKE1 cluster
func CreateRKE1Cluster(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.RKEKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, err := rke.CreateRKEMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	return nil
}
