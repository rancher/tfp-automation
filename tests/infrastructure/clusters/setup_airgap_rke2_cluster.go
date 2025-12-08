package clusters

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	registry "github.com/rancher/tfp-automation/framework/set/resources/registries/createRegistry"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateAirgappedRKE2Cluster is a function that creates an airgap RKE2 cluster, either via CLI or web application
func CreateAirgappedRKE2Cluster(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapRKE2KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{serverOne, serverTwo, serverThree}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	registryPublicIP := terraform.Output(t, terraformOptions, registryPublicIP)
	bastionPublicIP := terraform.Output(t, terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverTwoPrivateIP := terraform.Output(t, terraformOptions, serverTwoPrivateIP)
	serverThreePrivateIP := terraform.Output(t, terraformOptions, serverThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating registry...")
	file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, registryPublicIP, nonAuthRegistry)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating airgap RKE2 cluster...")
	file, err = rke2.CreateAirgapRKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP, registryPublicIP, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	return nil
}
