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
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/k3s"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/k3s/squid"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateProxyK3SCluster is a function that creates a K3S cluster with a proxy either via CLI or web application
func CreateProxyK3SCluster(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.ProxyK3SKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
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
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	bastionPublicDNS := terraform.Output(t, terraformOptions, bastionPublicDNS)
	bastionPrivateIP := terraform.Output(t, terraformOptions, bastionPrivateIP)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverTwoPrivateIP := terraform.Output(t, terraformOptions, serverTwoPrivateIP)
	serverThreePrivateIP := terraform.Output(t, terraformOptions, serverThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating squid proxy...")
	file, err = squid.CreateSquidProxy(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating K3S cluster...")
	file, err = k3s.CreateK3SCluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, bastionPrivateIP, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	return nil
}
