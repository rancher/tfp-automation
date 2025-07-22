package recurring

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	resources "github.com/rancher/tfp-automation/framework/set/resources/sanity"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/infrastructure"
	"github.com/rancher/tfp-automation/tests/rancher2/snapshot"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpRancher2RecurringRunsTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	cattleConfig               map[string]any
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
}

func (r *TfpRancher2RecurringRunsTestSuite) TearDownSuite() {
	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, r.terratestConfig.PathToRepo, r.terraformConfig.Provider)
	cleanup.Cleanup(r.T(), r.standaloneTerraformOptions, keyPath)
}

func (r *TfpRancher2RecurringRunsTestSuite) SetupSuite() {
	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig = config.LoadTFPConfigs(r.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, r.terratestConfig.PathToRepo, r.terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.standaloneTerraformOptions = standaloneTerraformOptions

	_, err := resources.CreateMainTF(r.T(), r.standaloneTerraformOptions, keyPath, r.rancherConfig, r.terraformConfig, r.terratestConfig)
	require.NoError(r.T(), err)

	testSession := session.NewSession()
	r.session = testSession

	client, err := infrastructure.PostRancherSetup(r.T(), r.rancherConfig, testSession, r.terraformConfig.Standalone.RancherHostname, false, false)
	require.NoError(r.T(), err)

	r.client = client

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.terraformOptions = terraformOptions
}

func (r *TfpRancher2RecurringRunsTestSuite) TestTfpRecurringProvisionCustomCluster() {
	tests := []struct {
		name   string
		module string
	}{
		{"Custom TFP RKE2", modules.CustomEC2RKE2},
		{"Custom TFP RKE2 Windows 2019", modules.CustomEC2RKE2Windows2019},
		{"Custom TFP RKE2 Windows 2022", modules.CustomEC2RKE2Windows2022},
		{"Custom TFP K3S", modules.CustomEC2K3s},
	}

	customClusterNames := []string{}
	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, customClusterNames)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusterIDs, _ = provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, true, true, true, customClusterNames)
				provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			}
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRancher2RecurringRunsTestSuite) TestTfpRecurringProvisionImportedCluster() {
	tests := []struct {
		name   string
		module string
	}{
		{"Upgrade Imported TFP RKE2", modules.ImportEC2RKE2},
		{"Upgrade Imported TFP RKE2 Windows 2019", modules.ImportEC2RKE2Windows2019},
		{"Upgrade Imported TFP RKE2 Windows 2022", modules.ImportEC2RKE2Windows2022},
		{"Upgrade Imported TFP K3S", modules.ImportEC2K3s},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " " + currentDate

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, true, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)

			err = imported.SetUpgradeImportedCluster(r.client, terraform)
			require.NoError(r.T(), err)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func (r *TfpRancher2RecurringRunsTestSuite) TestTfpRecurringSnapshotRestore() {
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	snapshotRestoreNone := config.TerratestConfig{
		SnapshotInput: config.Snapshots{
			SnapshotRestore: "none",
		},
	}

	tests := []struct {
		name         string
		module       string
		nodeRoles    []config.Nodepool
		etcdSnapshot config.TerratestConfig
	}{
		{"RKE2 Snapshot Restore", modules.EC2RKE2, nodeRolesDedicated, snapshotRestoreNone},
		{"K3S Snapshot Restore", modules.EC2K3s, nodeRolesDedicated, snapshotRestoreNone},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		configMap, err := provisioning.UniquifyTerraform([]map[string]any{r.cattleConfig})
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, r.client.RancherConfig.AdminToken, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "nodepools"}, tt.nodeRoles, configMap[0])
		require.NoError(r.T(), err)

		_, err = operations.ReplaceValue([]string{"terratest", "snapshotInput", "snapshotRestore"}, tt.etcdSnapshot.SnapshotInput.SnapshotRestore, configMap[0])
		require.NoError(r.T(), err)

		provisioning.GetK8sVersion(r.T(), r.client, r.terratestConfig, r.terraformConfig, configs.DefaultK8sVersion, configMap)

		rancher, terraform, terratest := config.LoadTFPConfigs(configMap[0])

		currentDate := time.Now().Format("2006-01-02 03:04PM")
		tt.name = tt.name + " Kubernetes version: " + terratest.KubernetesVersion + " " + currentDate

		r.Run(tt.name, func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs, _ := provisioning.Provision(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file, false, false, false, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)

			snapshot.RestoreSnapshot(r.T(), r.client, rancher, terraform, terratest, testUser, testPassword, r.terraformOptions, configMap, newFile, rootBody, file)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest(r.terratestConfig)
	}
}

func TestTfpRancher2RecurringRunsTestSuite(t *testing.T) {
	suite.Run(t, new(TfpRancher2RecurringRunsTestSuite))
}
