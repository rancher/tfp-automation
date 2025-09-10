package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	rke "github.com/rancher/tfp-automation/framework/set/resources/rke"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CreateRKE1ClusterTestSuite struct {
	suite.Suite
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (i *CreateRKE1ClusterTestSuite) TestCreateRKE1Cluster() {
	i.rancherConfig, i.terraformConfig, i.terratestConfig, _ = config.LoadTFPConfigs(i.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RKEKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	_, err := rke.CreateRKEMainTF(i.T(), i.terraformOptions, keyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)
}

func TestCreateRKE1ClusterTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRKE1ClusterTestSuite))
}
