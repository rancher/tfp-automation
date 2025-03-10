package os

import (
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations/permutations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"

	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/permutationsdata"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	amiBatchSize = 2
)

type OSValidationTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
	cattleConfig     map[string]any
	permutedConfigs  []map[string]any
}

func (p *OSValidationTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

	modulePermutation, err := permutationsdata.CreateModulePermutation(p.cattleConfig)
	require.NoError(p.T(), err)

	modulePermutation.KeyPathValueRelationships, err = permutationsdata.CreateAMIRelationships(p.cattleConfig)
	require.NoError(p.T(), err)

	k8sRelationships, err := permutationsdata.CreateK8sRelationships(p.cattleConfig)
	require.NoError(p.T(), err)

	modulePermutation.KeyPathValueRelationships = append(modulePermutation.KeyPathValueRelationships, k8sRelationships...)

	cniPermutation, err := permutationsdata.CreateCNIPermutation(p.cattleConfig)
	require.NoError(p.T(), err)

	permutedConfigs, err := permutations.Permute([]permutations.Permutation{*modulePermutation, *cniPermutation}, p.cattleConfig)
	require.NoError(p.T(), err)

	p.permutedConfigs, err = provisioning.UniquifyTerraform(permutedConfigs)
	require.NoError(p.T(), err)

	p.rancherConfig, p.terraformConfig, p.terratestConfig = config.LoadTFPConfigs(p.permutedConfigs[0])

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(p.T(), p.terraformConfig, p.terratestConfig, keyPath)
	p.terraformOptions = terraformOptions
}

func (p *OSValidationTestSuite) TestDynamicOSValidation() {
	//Batch configs so that not all AMIs run in parallel
	configBatches := map[string][]map[string]any{}
	for _, cattleConfig := range p.permutedConfigs {
		_, terraformConfig, _ := config.LoadTFPConfigs(cattleConfig)
		configBatches[terraformConfig.AWSConfig.AMI] = append(configBatches[terraformConfig.AWSConfig.AMI], cattleConfig)
	}

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)

	for ami, batch := range configBatches {
		testUser, testPassword := configs.CreateTestCredentials()

		var clusterIDs []string
		p.Run("Parallel_Provisioning_"+ami, func() {
			for _, cattleConfig := range batch {
				_, terraformConfig, terratestConfig := config.LoadTFPConfigs(cattleConfig)
				logrus.Infof("Provisioning Cluster Type: %s, "+"K8s Version: %s, "+"CNI: %s", terraformConfig.Module, terratestConfig.KubernetesVersion, terraformConfig.CNI)
			}

			clusterIDs = provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, p.terratestConfig, testUser, testPassword, p.terraformOptions, batch, false)
			time.Sleep(2 * time.Minute)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
		})

		workloadTests(&p.Suite, p.client, clusterIDs)
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestOSValidationTestSuite(t *testing.T) {
	suite.Run(t, new(OSValidationTestSuite))
}
