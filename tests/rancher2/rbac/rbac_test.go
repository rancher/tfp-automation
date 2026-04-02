//go:build validation || recurring

package rbac

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	nested "github.com/rancher/tfp-automation/tests/extensions/nestedModules"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rb "github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RBACTestSuite struct {
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

func (r *RBACTestSuite) SetupSuite() {
	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig, _ = config.LoadTFPConfigs(r.cattleConfig)

	testSession := session.NewSession()
	r.session = testSession

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)

	r.terraformOptions = terraformOptions

	client, err := ranchers.PostRancherSetup(r.T(), r.terraformOptions, r.rancherConfig, r.session, r.rancherConfig.Host, keyPath, false)
	require.NoError(r.T(), err)

	r.client = client
}

func (r *RBACTestSuite) TestTfpRBAC() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	r.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(r.client)
	require.NoError(r.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(r.T(), r.terraformOptions, r.rancherConfig, testUser, testPassword)
	require.NoError(r.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	rke2Module, _, _, k3sModule := provisioning.DownstreamClusterModules(r.terraformConfig)

	tests := []struct {
		name     string
		module   string
		rbacRole config.Role
	}{
		{"RKE2_Cluster_Owner", rke2Module, config.ClusterOwner},
		{"RKE2_Project_Owner", rke2Module, config.ProjectOwner},
		{"K3S_Cluster_Owner", k3sModule, config.ClusterOwner},
		{"K3S_Project_Owner", k3sModule, config.ProjectOwner},
	}

	for _, tt := range tests {
		r.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(r.terraformConfig, r.terratestConfig, r.terraformOptions, tt.name, "/modules/rancher2")
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(r.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(r.T(), err)

			_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, nodeRolesDedicated, cattleConfig)
			require.NoError(r.T(), err)

			provisioning.GetK8sVersion(r.client, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), perTestTerraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, r.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, false, clusterIDs, nil, nestedRancherModuleDir)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(r.T(), r.client, clusterIDs)

			cluster, err := r.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(r.T(), err)

			err = pods.VerifyClusterPods(r.client, cluster)
			require.NoError(r.T(), err)

			rb.RBAC(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, tt.rbacRole, newFile, rootBody, file, nestedRancherModuleDir)

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func TestTfpRBACTestSuite(t *testing.T) {
	suite.Run(t, new(RBACTestSuite))
}
