//go:build validation || recurring

package provisioning

import (
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
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
	"github.com/rancher/tfp-automation/defaults/stevetypes"
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

type ProvisionCustomTestSuite struct {
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

func (p *ProvisionCustomTestSuite) SetupSuite() {
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

func (p *ProvisionCustomTestSuite) TestTfpProvisionCustom() {
	var err error
	var testUser, testPassword string
	var clusterIDs []string

	customClusterNames := []string{}

	p.standardUserClient, testUser, testPassword, err = standarduser.CreateStandardUser(p.client)
	require.NoError(p.T(), err)

	standardUserToken, err := ranchers.CreateStandardUserToken(p.T(), p.terraformOptions, p.rancherConfig, testUser, testPassword)
	require.NoError(p.T(), err)

	standardToken := standardUserToken.Token

	tests := []struct {
		name   string
		module string
	}{
		{"Custom_TFP_RKE2", modules.CustomEC2RKE2},
		{"Custom_TFP_RKE2_Windows_2019", modules.CustomEC2RKE2Windows2019},
		{"Custom_TFP_RKE2_Windows_2022", modules.CustomEC2RKE2Windows2022},
		{"Custom_TFP_K3S", modules.CustomEC2K3s},
	}

	for _, tt := range tests {
		p.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.terraformConfig, p.terratestConfig, p.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			cattleConfig, err := provisioning.UniquifyTerraform(p.cattleConfig)
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, cattleConfig)
			require.NoError(p.T(), err)

			_, err = operations.ReplaceValue([]string{"terraform", "module"}, tt.module, cattleConfig)
			require.NoError(p.T(), err)

			provisioning.GetK8sVersion(p.client, cattleConfig)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(cattleConfig)

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), perTestTerraformOptions, keyPath)

			clusters, customClusterNames := provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
			require.NoError(p.T(), err)
			err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
			require.NoError(p.T(), err)
			cluster, err := p.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(p.T(), err)

			err = pods.VerifyClusterPods(p.client, cluster)
			require.NoError(p.T(), err)

			if strings.Contains(terraform.Module, clustertypes.WINDOWS) {
				clusters, _ = provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, []map[string]any{cattleConfig}, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
				err = provisioningActions.VerifyClusterReady(p.client, clusters[0])
				require.NoError(p.T(), err)
				err = clusterActions.VerifyServiceAccountTokenSecret(p.client, clusters[0].Name)
				require.NoError(p.T(), err)
				err = pods.VerifyClusterPods(p.client, cluster)
				require.NoError(p.T(), err)
			}

			params := tfpQase.GetProvisioningSchemaParams(cattleConfig)
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

func TestTfpProvisionCustomTestSuite(t *testing.T) {
	suite.Run(t, new(ProvisionCustomTestSuite))
}
