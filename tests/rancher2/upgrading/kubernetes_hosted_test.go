//go:build validation || recurring

package upgrading

import (
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
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

type KubernetesUpgradeHostedTestSuite struct {
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

func (k *KubernetesUpgradeHostedTestSuite) SetupSuite() {
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

func (k *KubernetesUpgradeHostedTestSuite) TestTfpKubernetesUpgradeHosted() {
	var err error
	var testUser, testPassword string

	k.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(k.client)
	require.NoError(k.T(), err)

	aksNodePools := []config.Nodepool{{Quantity: 3}}
	eksNodePools := []config.Nodepool{{DiskSize: 100, InstanceType: k.terraformConfig.AWSConfig.AWSInstanceType, DesiredSize: 3, MaxSize: 3, MinSize: 3}}
	gkeNodePools := []config.Nodepool{{Quantity: 3, MaxPodsConstraint: 110}}

	tests := []struct {
		name                      string
		module                    string
		nodePools                 []config.Nodepool
		kubernetesVersion         string
		upgradedKubernetesVersion string
	}{
		{"Upgrade_AKS_Cluster", modules.AKS, aksNodePools, k.terratestConfig.AKSKubernetesVersion, k.terratestConfig.UpgradedAKSKubernetesVersion},
		{"Upgrade_EKS_Cluster", modules.EKS, eksNodePools, k.terratestConfig.EKSKubernetesVersion, k.terratestConfig.UpgradedEKSKubernetesVersion},
		{"Upgrade_GKE_Cluster", modules.GKE, gkeNodePools, k.terratestConfig.GKEKubernetesVersion, k.terratestConfig.UpgradedGKEKubernetesVersion},
	}

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(k.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{k.cattleConfig})
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodePools, configMap[0])
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "kubernetesVersion"}, tt.kubernetesVersion, configMap[0])
		require.NoError(k.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "upgradedKubernetesVersion"}, tt.upgradedKubernetesVersion, configMap[0])
		require.NoError(k.T(), err)

		rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

		k.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, k.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(k.T(), k.terraformOptions, keyPath)

			adminClient, err := provisioning.FetchAdminClient(k.T(), k.client)
			require.NoError(k.T(), err)

			clusterIDs, _ := provisioning.Provision(k.T(), k.client, k.standardUserClient, rancher, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(k.T(), adminClient, clusterIDs)
			provisioning.VerifyV3ClustersPods(k.T(), adminClient, clusterIDs)

			provisioning.KubernetesUpgrade(k.T(), k.client, k.standardUserClient, rancher, terraform, terratest, testUser, testPassword, k.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			time.Sleep(4 * time.Minute)
			provisioning.VerifyClustersState(k.T(), adminClient, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(k.T(), adminClient, clusterIDs)
			provisioning.VerifyV3ClustersPods(k.T(), adminClient, clusterIDs)
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

func TestTfpKubernetesUpgradeHostedTestSuite(t *testing.T) {
	suite.Run(t, new(KubernetesUpgradeHostedTestSuite))
}
