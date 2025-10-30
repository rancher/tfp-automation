//go:build dynamic

package upgrading

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DynamicKubernetesUpgradeTestSuite struct {
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

func (k *DynamicKubernetesUpgradeTestSuite) SetupSuite() {
	testSession := session.NewSession()
	k.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(k.T(), err)

	k.client = client

	k.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	k.rancherConfig, k.terraformConfig, k.terratestConfig, _ = config.LoadTFPConfigs(k.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(k.T(), k.terraformConfig, k.terratestConfig, keyPath)
	k.terraformOptions = terraformOptions
}

func (k *DynamicKubernetesUpgradeTestSuite) TestTfpKubernetesUpgradeDynamicInput() {
	var err error
	var testUser, testPassword string

	k.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(k.client)
	require.NoError(k.T(), err)

	tests := []struct {
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(k.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{k.cattleConfig})
		require.NoError(k.T(), err)

		provisioning.GetK8sVersion(k.T(), k.client, k.terratestConfig, k.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		k.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(k.T(), k.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(k.T(), k.client)
			require.NoError(k.T(), err)

			clusterIDs, _ := provisioning.Provision(k.T(), k.client, k.standardUserClient, rancher, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(k.T(), adminClient, clusterIDs)

			cluster, err := k.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(k.T(), err)

			pods.VerifyClusterPods(k.T(), adminClient, cluster)

			provisioning.KubernetesUpgrade(k.T(), k.client, k.standardUserClient, rancher, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(k.T(), adminClient, clusterIDs)
			pods.VerifyClusterPods(k.T(), adminClient, cluster)
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

func TestDynamicTfpKubernetesUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(DynamicKubernetesUpgradeTestSuite))
}
