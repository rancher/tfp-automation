package rke

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	rke "github.com/rancher/tfp-automation/framework/set/resources/rke"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/stretchr/testify/suite"
)

type RKEProviderTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (t *RKEProviderTestSuite) TearDownSuite() {
	keyPath := rke.KeyPath()
	cleanup.Cleanup(t.T(), t.terraformOptions, keyPath)
}

func (t *RKEProviderTestSuite) TestCreateRKECluster() {
	t.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, t.terraformConfig)

	t.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, t.terratestConfig)

	keyPath := rke.KeyPath()
	terraformOptions := framework.Setup(t.T(), t.terraformConfig, t.terratestConfig, keyPath)
	t.terraformOptions = terraformOptions

	rke.CreateRKEMainTF(t.T(), t.terraformOptions, keyPath, t.terraformConfig, t.terratestConfig)

	if t.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestRKEProviderTestSuite(t *testing.T) {
	suite.Run(t, new(RKEProviderTestSuite))
}
