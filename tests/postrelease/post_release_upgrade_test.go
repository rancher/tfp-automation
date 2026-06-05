package postrelease

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"

	setupstandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/standard"
	upgradestandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/standard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type TfpPostReleaseUpgradeTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	upgradeTerraformOptions    *terraform.Options
	terraformOptions           *terraform.Options
	serverNodeOne              string
}

func (s *TfpPostReleaseUpgradeTestSuite) TestTfpPostReleaseUpgrade() {
	tests := []struct {
		name string
	}{
		{"Post_Release_Upgrade"},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			testSession := session.NewSession()
			s.session = testSession
			s.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

			s.client, s.serverNodeOne, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = setupstandard.SetupRancher(s.T(), s.session, keypath.SanityKeyPath, s.cattleConfig)
			s.client, s.cattleConfig, s.terraformOptions, s.upgradeTerraformOptions = upgradestandard.UpgradeRancher(s.T(), s.client, s.serverNodeOne, s.session, s.cattleConfig)
			_, s.terraformConfig, s.terratestConfig, _ = config.LoadTFPConfigs(s.cattleConfig)

			params := tfpQase.GetProvisioningSchemaParams(s.terraformConfig, s.terratestConfig)
			err := qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}
}

func TestTfpPostReleaseUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(TfpPostReleaseUpgradeTestSuite))
}
