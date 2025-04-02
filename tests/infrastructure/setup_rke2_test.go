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
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateRKE2ClusterTestSuite struct {
	suite.Suite
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

const (
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"

	rke2ServerOnePublicIP = "rke2_server1_public_ip"
	registryPublicIP      = "registry_public_ip"

	nonAuthRegistry = "non_auth_registry"

	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPublicIP    = "rke2_server2_public_ip"
	rke2ServerThreePublicIP  = "rke2_server3_public_ip"
	rke2BastionPublicIP      = "rke2_bastion_public_ip"
	rke2ServerTwoPrivateIP   = "rke2_server2_private_ip"
	rke2ServerThreePrivateIP = "rke2_server3_private_ip"

	terraformConst = "terraform"
)

func (i *CreateRKE2ClusterTestSuite) TestCreateRKE2Cluster() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)

	keyPath := rancher2.SetKeyPath(keypath.RKE2KeyPath, i.terraformConfig)
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

	providerTunnel := providers.TunnelToProvider(i.terraformConfig)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, i.terraformConfig, i.terratestConfig, instances)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	rke2ServerOnePublicIP := terraform.Output(i.T(), terraformOptions, rke2ServerOnePublicIP)
	rke2ServerOnePrivateIP := terraform.Output(i.T(), terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPublicIP := terraform.Output(i.T(), terraformOptions, rke2ServerTwoPublicIP)
	rke2ServerThreePublicIP := terraform.Output(i.T(), terraformOptions, rke2ServerThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, i.terraformConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP)
	require.NoError(i.T(), err)

	terraform.InitAndApply(i.T(), terraformOptions)

	logrus.Infof("Kubeconfig file is located in /home/%s/.kube in the node: %s", i.terraformConfig.Standalone.OSUser, rke2ServerOnePublicIP)
}

func TestCreateRKE2ClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRKE2ClusterTestSuite))
}
