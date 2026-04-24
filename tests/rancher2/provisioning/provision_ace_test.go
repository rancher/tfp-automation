//go:build validation || recurring

package provisioning

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProvisionACETestSuite struct {
	suite.Suite
	client             *rancher.Client
	standardUserClient *rancher.Client
	session            *session.Session
	cattleConfig       map[string]any
	rancherConfig      *rancher.Config
	terraformConfig    *config.TerraformConfig
	terratestConfig    *config.TerratestConfig
	terraformOptions   *terraform.Options
}

func (p *ProvisionACETestSuite) SetupSuite() {
	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	p.rancherConfig, p.terraformConfig, p.terratestConfig, _ = config.LoadTFPConfigs(p.cattleConfig)

	testSession := session.NewSession()
	p.session = testSession

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)

	p.terraformOptions = terraformOptions

	client, err := ranchers.PostRancherSetup(p.T(), p.terraformOptions, p.rancherConfig, p.session, p.rancherConfig.Host, keyPath, false)
	require.NoError(p.T(), err)

	p.client = client
}

func (p *ProvisionACETestSuite) TestTfpProvisionACE() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	p.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(p.client)
	require.NoError(p.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(p.T(), p.terraformOptions, p.rancherConfig, testUser, testPassword)
	require.NoError(p.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	rke2Module, _, _, k3sModule := provisioning.DownstreamClusterModules(p.terraformConfig)

	localAuthEndpoint := config.TerraformConfig{
		LocalAuthEndpoint: true,
	}

	tests := []struct {
		name         string
		module       string
		nodeRoles    []config.Nodepool
		authEndpoint config.TerraformConfig
	}{
		{"RKE2_ACE", rke2Module, nodeRolesDedicated, localAuthEndpoint},
		{"K3S_ACE", k3sModule, nodeRolesDedicated, localAuthEndpoint},
	}

	for _, tt := range tests {
		p.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(p.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.LocalAuthEndpoint = tt.authEndpoint.LocalAuthEndpoint
			terraform.Module = tt.module

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.terraformConfig, p.terratestConfig, p.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(p.client, terraform, terratest)
			require.NoError(p.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), perTestTerraformOptions, keyPath)

			clusters, _ := provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
			require.NoError(p.T(), err)

			err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
			require.NoError(p.T(), err)

			err = pods.VerifyClusterPods(p.client, clusters[0])
			require.NoError(p.T(), err)

			provisioningActions.VerifyACE(p.T(), p.client, clusters[0])

			params := tfpQase.GetProvisioningSchemaParams(p.terraformConfig, p.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if p.terratestConfig.LocalQaseReporting {
		results.ReportTest(p.terratestConfig)
	}
}

func TestProvisionACETestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionACETestSuite))
}
