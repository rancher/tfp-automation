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
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/k3s"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateIPv6K3SCluster is a function that creates an IPv6 K3S cluster either via CLI or web application
func CreateIPv6K3SCluster(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.IPv6RKE2K3SKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{bastion}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateIPv6(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	bastionPublicIP := terraform.Output(t, terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverOnePublicIP := terraform.Output(t, terraformOptions, serverOnePublicIP)
	serverTwoPrivateIP := terraform.Output(t, terraformOptions, serverTwoPrivateIP)
	serverTwoPublicIP := terraform.Output(t, terraformOptions, serverTwoPublicIP)
	serverThreePrivateIP := terraform.Output(t, terraformOptions, serverThreePrivateIP)
	serverThreePublicIP := terraform.Output(t, terraformOptions, serverThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating K3S cluster...")
	file, err = k3s.CreateIPv6K3SCluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP, serverOnePublicIP, serverTwoPublicIP, serverThreePublicIP,
		serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	return nil
}
