package rbac

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rb "github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RBACTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (r *RBACTestSuite) SetupSuite() {
	testSession := session.NewSession()
	r.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig = config.LoadTFPConfigs(r.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.terraformOptions = terraformOptions
}

func (r *RBACTestSuite) TestTfpRBAC() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name     string
		rbacRole config.Role
	}{
		{"Cluster Owner", config.ClusterOwner},
		{"Project Owner", config.ProjectOwner},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF()
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, nodeRolesDedicated, configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		_, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		tt.name = tt.name + " Module: " + r.terraformConfig.Module

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(r.T(), r.client)
			require.NoError(r.T(), err)

			clusterIDs, _ := provisioning.Provision(r.T(), adminClient, terraform, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(r.T(), adminClient, clusterIDs)
			rb.RBAC(r.T(), r.client, r.rancherConfig, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, tt.rbacRole, newFile, rootBody, file)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpRBACTestSuite(t *testing.T) {
	suite.Run(t, new(RBACTestSuite))
}
