package clusters

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/k3s"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateIPv6K3SClusterTestSuite struct {
	suite.Suite
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *CreateIPv6K3SClusterTestSuite) TestCreateIPv6K3SCluster() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)
	_, keyPath := rancher2.SetKeyPath(keypath.IPv6RKE2K3SKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{bastion}

	providerTunnel := providers.TunnelToProvider(i.terraformConfig.Provider)
	file, err := providerTunnel.CreateIPv6(file, newFile, tfBlockBody, rootBody, i.terraformConfig, i.terratestConfig, instances)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	bastionPublicIP := terraform.Output(i.T(), terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.Output(i.T(), terraformOptions, serverOnePrivateIP)
	serverOnePublicIP := terraform.Output(i.T(), terraformOptions, serverOnePublicIP)
	serverTwoPrivateIP := terraform.Output(i.T(), terraformOptions, serverTwoPrivateIP)
	serverTwoPublicIP := terraform.Output(i.T(), terraformOptions, serverTwoPublicIP)
	serverThreePrivateIP := terraform.Output(i.T(), terraformOptions, serverThreePrivateIP)
	serverThreePublicIP := terraform.Output(i.T(), terraformOptions, serverThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating K3S cluster...")
	file, err = k3s.CreateIPv6K3SCluster(file, newFile, rootBody, i.terraformConfig, i.terratestConfig, bastionPublicIP, serverOnePublicIP, serverTwoPublicIP, serverThreePublicIP,
		serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)
}

func TestCreateIPv6K3SClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateIPv6K3SClusterTestSuite))
}
