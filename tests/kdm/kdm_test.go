package kdm

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/shepherd/pkg/session"
	clusterActions "github.com/rancher/tests/actions/clusters"
	provisioningActions "github.com/rancher/tests/actions/provisioning"

	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
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

type KDMTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	standardUserClient         *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneConfig           *config.Standalone
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (k *KDMTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, k.terratestConfig.PathToRepo, k.terraformConfig.Provider)
	cleanup.Cleanup(k.T(), k.standaloneTerraformOptions, keyPath)
}

func (k *KDMTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	k.client, _, k.standaloneTerraformOptions, k.terraformOptions, k.cattleConfig = ranchers.SetupRancher(k.T(), k.session, keypath.SanityKeyPath)
	k.rancherConfig, k.terraformConfig, k.terratestConfig, k.standaloneConfig = config.LoadTFPConfigs(k.cattleConfig)
}

func (k *KDMTestSuite) TestKDM() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	rawKdm, err := provisioning.FetchSetting(k.client, "rke-metadata-config")
	require.NoError(k.T(), err)

	if strings.HasPrefix(rawKdm, "{") {
		var m map[string]string
		err := json.Unmarshal([]byte(rawKdm), &m)
		require.NoError(k.T(), err)
		rawKdm = m["url"]
	}

	kdmURL := rawKdm

	kdmVersions := provisioning.VerifyKDMUrl(k.T(), kdmURL, k.standaloneConfig.RancherTagVersion)

	rke2Versions, err := kubernetesversions.ListRKE2AllVersions(k.client)
	require.NoError(k.T(), err)

	k3sVersions, err := kubernetesversions.ListK3SAllVersions(k.client)
	require.NoError(k.T(), err)

	provisioning.VerifyKDMVersions(k.T(), kdmVersions, rke2Versions, clustertypes.RKE2)
	provisioning.VerifyKDMVersions(k.T(), kdmVersions, k3sVersions, clustertypes.K3S)

	k.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(k.client)
	require.NoError(k.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(k.T(), k.terraformOptions, k.rancherConfig, testUser, testPassword)
	require.NoError(k.T(), err)

	standardToken := standardUserToken.Token

	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		nodeRoles []config.Nodepool
		module    string
	}{
		{"KDM_Sanity_RKE2", nodeRolesDedicated, modules.NodeDriverAWSRKE2},
		{"KDM_Sanity_K3S", nodeRolesDedicated, modules.NodeDriverAWSK3S},
	}

	for _, tt := range tests {
		k.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(k.cattleConfig)
			rancher.AdminToken = standardToken
			terratest.Nodepools = tt.nodeRoles
			terraform.Module = tt.module

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(k.terraformConfig, k.terratestConfig, k.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()
			terratest, err = provisioning.GetK8sVersion(k.standardUserClient, terraform, terratest)
			require.NoError(k.T(), err)

			terraform = provisioning.UniquifyTerraform(terraform)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(k.T(), perTestTerraformOptions, keyPath)

			logrus.Infof("Provisioning cluster (%s)", terraform.ResourcePrefix)
			clusters, _ := provisioning.Provision(k.T(), k.client, k.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, newFile, rootBody, file, false, false, true, clusterIDs, "", nestedRancherModuleDir)

			logrus.Infof("Verifying the cluster is ready (%s)", clusters[0].Name)
			err = provisioningActions.VerifyClusterReady(k.client, clusters[0])
			require.NoError(k.T(), err)

			logrus.Infof("Verifying service account token secret (%s)", clusters[0].Name)
			err = clusterActions.VerifyServiceAccountTokenSecret(k.client, clusters[0].Name)
			require.NoError(k.T(), err)

			logrus.Infof("Verifying cluster pods (%s)", clusters[0].Name)
			err = pods.VerifyClusterPods(k.client, clusters[0])
			require.NoError(k.T(), err)

			params := tfpQase.GetProvisioningSchemaParams(k.terraformConfig, k.terratestConfig)
			err = qase.UpdateSchemaParameters(tt.name, params)
			if err != nil {
				logrus.Warningf("Failed to upload schema parameters %s", err)
			}
		})
	}

	if k.terratestConfig.LocalQaseReporting {
		results.ReportTest(k.terratestConfig)
	}
}

func TestKDMTestSuite(t *testing.T) {
	suite.Run(t, new(KDMTestSuite))
}
