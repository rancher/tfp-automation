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

	keyPath := rancher2.SetKeyPath(keypath.AirgapRKE2KeyPath, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}

	providerTunnel := providers.TunnelToProvider(i.terraformConfig.Provider)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, i.terraformConfig, i.terratestConfig, instances)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	registryPublicIP := terraform.Output(i.T(), terraformOptions, registryPublicIP)
	rke2BastionPublicIP := terraform.Output(i.T(), terraformOptions, rke2BastionPublicIP)
	rke2ServerOnePrivateIP := terraform.Output(i.T(), terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPrivateIP := terraform.Output(i.T(), terraformOptions, rke2ServerTwoPrivateIP)
	rke2ServerThreePrivateIP := terraform.Output(i.T(), terraformOptions, rke2ServerThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating registry...")
	file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, i.terraformConfig, registryPublicIP, nonAuthRegistry)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating airgap RKE2 cluster...")
	file, err = rke2.CreateAirgapRKE2Cluster(file, newFile, rootBody, i.terraformConfig, rke2BastionPublicIP, registryPublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	logrus.Infof("Kubeconfig file is located in /home/%s/.kube in the bastion node: %s", i.terraformConfig.Standalone.OSUser, rke2BastionPublicIP)
}

func TestCreateAirgappedRKE2ClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAirgappedRKE2ClusterTestSuite))
}
