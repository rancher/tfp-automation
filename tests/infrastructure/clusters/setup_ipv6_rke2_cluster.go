package clusters

import (
	"context"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateIPv6RKE2Cluster is a function that creates an IPv6 RKE2 cluster either via CLI or web application
func CreateIPv6RKE2Cluster(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.IPv6RKE2KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
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

	terraform.InitAndApplyContext(t, context.Background(), terraformOptions)

	bastionPublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverOnePrivateIP)
	serverOnePublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverOnePublicIP)
	serverTwoPrivateIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverTwoPrivateIP)
	serverTwoPublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverTwoPublicIP)
	serverThreePrivateIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverThreePrivateIP)
	serverThreePublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateIPv6RKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP, serverOnePublicIP, serverTwoPublicIP, serverThreePublicIP,
		serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(t, err)

	terraform.InitAndApplyContext(t, context.Background(), terraformOptions)

	return nil
}
