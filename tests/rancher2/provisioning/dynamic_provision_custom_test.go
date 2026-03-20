//go:build dynamic

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
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
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

type DynamicProvisionCustomTestSuite struct {
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

func (p *DynamicProvisionCustomTestSuite) SetupSuite() {
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

func (p *DynamicProvisionCustomTestSuite) TestTfpProvisionCustomDynamicInput() {
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
		name string
	}{
		{config.StandardClientName.String()},
	}

	for _, tt := range tests {
		p.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nestedRancherModuleDir, perTestTerraformOptions, err := nested.CreateNestedModules(p.terraformConfig, p.terratestConfig, p.terraformOptions, tt.name, configs.NestedRancherModuleDir)
			require.NoError(t, err)
			defer os.RemoveAll(nestedRancherModuleDir)

			newFile, rootBody, file := rancher2.InitializeNestedMainTFs(nestedRancherModuleDir)
			defer file.Close()

			configMap, err := provisioning.UniquifyTerraform([]map[string]any{p.cattleConfig})
			require.NoError(t, err)

			_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, standardToken, configMap[0])
			require.NoError(p.T(), err)

			provisioning.GetK8sVersion(p.T(), p.client, p.terratestConfig, p.terraformConfig, configMap)

			rancher, terraform, terratest, _ := config.LoadTFPConfigs(configMap[0])

			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, p.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(p.T(), perTestTerraformOptions, keyPath)

			clusterIDs, customClusterNames := provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, configMap, newFile, rootBody, file, false, false, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
			provisioning.VerifyServiceAccountTokenSecret(p.T(), p.client, clusterIDs)

			cluster, err := p.client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetDefault + "/" + terraform.ResourcePrefix)
			require.NoError(p.T(), err)

			pods.VerifyClusterPods(p.client, cluster)

			if strings.Contains(p.terraformConfig.Module, clustertypes.WINDOWS) {
				clusterIDs, _ = provisioning.Provision(p.T(), p.client, p.standardUserClient, rancher, terraform, terratest, testUser, testPassword, perTestTerraformOptions, configMap, newFile, rootBody, file, true, true, true, clusterIDs, customClusterNames, nestedRancherModuleDir)
				provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
				provisioning.VerifyServiceAccountTokenSecret(p.T(), p.client, clusterIDs)
				pods.VerifyClusterPods(p.client, cluster)
			}

			params := tfpQase.GetProvisioningSchemaParams(configMap[0])
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

func TestDynamicProvisionCustomTestSuite(t *testing.T) {
	suite.Run(t, new(DynamicProvisionCustomTestSuite))
}
