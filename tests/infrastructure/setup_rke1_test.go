package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	rke "github.com/rancher/tfp-automation/framework/set/resources/rke"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateRKE1ClusterTestSuite struct {
	suite.Suite
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *CreateRKE1ClusterTestSuite) TestCreateRKE1Cluster() {
	i.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, i.terraformConfig)

	i.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, i.terratestConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RKEKeyPath, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	rkeNodeOne, err := rke.CreateRKEMainTF(i.T(), i.terraformOptions, keyPath, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	logrus.Infof("Kubeconfig file is located in /home/%s/.kube in the node: %s", i.terraformConfig.Standalone.OSUser, rkeNodeOne)
}

func TestCreateRKE1ClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRKE1ClusterTestSuite))
}
