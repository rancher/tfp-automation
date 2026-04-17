package postrelease

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpPostReleaseTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	standardUserClient         *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (s *TfpPostReleaseTestSuite) TestTfpPostRelease() {
	tests := []struct {
		name string
	}{
		{"Post_Release"},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			testSession := session.NewSession()
			s.session = testSession

			s.client, _, s.standaloneTerraformOptions, s.terraformOptions, s.cattleConfig = ranchers.SetupRancher(s.T(), s.session, keypath.SanityKeyPath)

			cluster, err := s.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetLocal + "/local")
			require.NoError(s.T(), err)

			err = pods.VerifyClusterPods(s.client, cluster)
			require.NoError(s.T(), err)

			params := tfpQase.GetProvisioningSchemaParams(s.terraformConfig, s.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}
}

func TestTfpPostReleaseTestSuite(t *testing.T) {
	suite.Run(t, new(TfpPostReleaseTestSuite))
}
