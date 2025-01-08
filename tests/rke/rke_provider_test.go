package rke

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup/rke"
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
	cleanup.ConfigRKECleanup(t.T(), t.terraformOptions)
}

func (t *RKEProviderTestSuite) TestCreateRKECluster() {
	t.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, t.terraformConfig)

	t.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, t.terratestConfig)

	terraformOptions, keyPath := framework.RKESetup(t.T(), t.terraformConfig, t.terratestConfig)
	t.terraformOptions = terraformOptions

	rke.CreateRKEMainTF(t.T(), t.terraformOptions, keyPath, t.terraformConfig, t.terratestConfig)

	if t.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestRKEProviderTestSuite(t *testing.T) {
	suite.Run(t, new(RKEProviderTestSuite))
}
