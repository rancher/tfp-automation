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
	"github.com/rancher/tfp-automation/framework/set/resources/dualstack/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateDualStackRKE2ClusterTestSuite struct {
	suite.Suite
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *CreateDualStackRKE2ClusterTestSuite) TestCreateDualStackRKE2Cluster() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)
	_, keyPath := rancher2.SetKeyPath(keypath.DualStackRKE2KeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
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

	rke2ServerOnePublicIP := terraform.Output(i.T(), terraformOptions, rke2ServerOnePublicIP)
	rke2ServerOnePrivateIP := terraform.Output(i.T(), terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPublicIP := terraform.Output(i.T(), terraformOptions, rke2ServerTwoPublicIP)
	rke2ServerThreePublicIP := terraform.Output(i.T(), terraformOptions, rke2ServerThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, i.terraformConfig, i.terratestConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)
}

func TestCreateDualStackRKE2ClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateDualStackRKE2ClusterTestSuite))
}
