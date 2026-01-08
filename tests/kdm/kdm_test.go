package kdm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters/kubernetesversions"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
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

	customClusterNames := []string{}

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
		{"KDM_Sanity_RKE2", nodeRolesDedicated, modules.EC2RKE2},
		{"KDM_Sanity_K3S", nodeRolesDedicated, modules.EC2K3s},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(k.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{k.cattleConfig})
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(k.T(), err)

		provisioning.GetK8sVersion(k.T(), k.standardUserClient, k.terratestConfig, k.terraformConfig, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		k.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(k.T(), k.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(k.T(), k.client, k.standardUserClient, rancher, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames)
			provisioning.VerifyClustersState(k.T(), k.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(k.T(), k.client, clusterIDs)

			cluster, err := k.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(k.T(), err)

			err = pods.VerifyClusterPods(k.client, cluster)
			require.NoError(k.T(), err)
		})

		params := tfpQase.GetProvisioningSchemaParams(configMap[0])
		err = qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if k.terratestConfig.LocalQaseReporting {
		results.ReportTest(k.terratestConfig)
	}
}

func TestKDMTestSuite(t *testing.T) {
	suite.Run(t, new(KDMTestSuite))
}
