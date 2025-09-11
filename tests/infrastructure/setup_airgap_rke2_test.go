package infrastructure

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
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
	"github.com/stretchr/testify/suite"
)

type CreateAirgappedRKE2ClusterTestSuite struct {
	suite.Suite
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *CreateAirgappedRKE2ClusterTestSuite) TestCreateAirgappedRKE2Cluster() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapRKE2KeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{serverOne, serverTwo, serverThree}

	providerTunnel := providers.TunnelToProvider(i.terraformConfig.Provider)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, i.terraformConfig, i.terratestConfig, instances)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	registryPublicIP := terraform.Output(i.T(), terraformOptions, registryPublicIP)
	bastionPublicIP := terraform.Output(i.T(), terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.Output(i.T(), terraformOptions, serverOnePrivateIP)
	serverTwoPrivateIP := terraform.Output(i.T(), terraformOptions, serverTwoPrivateIP)
	serverThreePrivateIP := terraform.Output(i.T(), terraformOptions, serverThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating registry...")
	file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, i.terraformConfig, i.terratestConfig, registryPublicIP, nonAuthRegistry)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating airgap RKE2 cluster...")
	file, err = rke2.CreateAirgapRKE2Cluster(file, newFile, rootBody, i.terraformConfig, i.terratestConfig, bastionPublicIP, registryPublicIP, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)
}

func TestCreateAirgappedRKE2ClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAirgappedRKE2ClusterTestSuite))
}
