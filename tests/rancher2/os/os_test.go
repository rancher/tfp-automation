package os

import (
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/cloudcredentials"
	clusterExtensions "github.com/rancher/shepherd/extensions/clusters"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/config/operations/permutations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/nodes/ec2"
	"github.com/rancher/tests/actions/workloads/cronjob"
	"github.com/rancher/tests/actions/workloads/daemonset"
	"github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tests/actions/workloads/statefulset"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/permutationsdata"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
	awsCredentials   cloudcredentials.AmazonEC2CredentialConfig
}

func (p *OSValidationTestSuite) SetupSuite() {
	testSession := session.NewSession()
	p.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(p.T(), err)

	p.client = client

	p.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

	p.cattleConfig, err = config.LoadPackageDefaults(p.cattleConfig, "")
	require.NoError(p.T(), err)

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

	_, terraformConfig, terratestConfig := config.LoadTFPConfigs(p.permutedConfigs[0])

	p.awsCredentials = cloudcredentials.AmazonEC2CredentialConfig{
		AccessKey:     terraformConfig.AWSCredentials.AWSAccessKey,
		SecretKey:     terraformConfig.AWSCredentials.AWSSecretKey,
		DefaultRegion: terraformConfig.AWSConfig.Region,
	}

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
	terraformOptions := framework.Setup(p.T(), terraformConfig, terratestConfig, keyPath)
	p.terraformOptions = terraformOptions
}

func (p *OSValidationTestSuite) TestDynamicOSValidation() {
	//Batch configs so that not all AMIs run in parallel
	configBatches := map[string][]map[string]any{}
	for _, cattleConfig := range p.permutedConfigs {
		_, terraformConfig, _ := config.LoadTFPConfigs(cattleConfig)
		configBatches[terraformConfig.AWSConfig.AMI] = append(configBatches[terraformConfig.AWSConfig.AMI], cattleConfig)
	}

	newFile, rootBody, file := rancher2.InitializeMainTF()
	defer file.Close()

	customClusterNames := []string{}

	for ami, batch := range configBatches {
		_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, "")
		defer cleanup.Cleanup(p.T(), p.terraformOptions, keyPath)
		testUser, testPassword := configs.CreateTestCredentials()

		var clusterIDs []string

		amiInfo, err := ec2.GetAMI(p.client, &p.awsCredentials, ami)
		require.NoError(p.T(), err)

		testName := "Parallel_Provisioning_" + *amiInfo.Images[0].Name
		p.Run(testName, func() {
			for _, cattleConfig := range batch {
				cattleConfig, err = operations.ReplaceValue([]string{"terraform", "awsConfig", "awsVolumeType"}, *amiInfo.Images[0].BlockDeviceMappings[0].Ebs.VolumeType, cattleConfig)
				require.NoError(p.T(), err)

				_, terraformConfig, terratestConfig := config.LoadTFPConfigs(cattleConfig)

				logrus.Infof("Provisioning Cluster Type: %s, "+"K8s Version: %s, "+"CNI: %s", terraformConfig.Module, terratestConfig.KubernetesVersion, terraformConfig.CNI)
			}

			clusterIDs, _ = provisioning.Provision(p.T(), p.client, p.rancherConfig, p.terraformConfig, testUser, testPassword, p.terraformOptions, batch, newFile, rootBody, file, false, false, true, customClusterNames)
			time.Sleep(2 * time.Minute)
			provisioning.VerifyClustersState(p.T(), p.client, clusterIDs)
		})

		workloadTests := []struct {
			name           string
			validationFunc func(client *rancher.Client, clusterID string) error
		}{
			{"WorkloadDeployment", deployment.VerifyCreateDeployment},
			{"WorkloadSideKick", deployment.VerifyCreateDeploymentSideKick},
			{"WorkloadDaemonSet", daemonset.VerifyCreateDaemonSet},
			{"WorkloadCronjob", cronjob.VerifyCreateCronjob},
			{"WorkloadStatefulset", statefulset.VerifyCreateStatefulset},
			{"WorkloadUpgrade", deployment.VerifyDeploymentUpgradeRollback},
			{"WorkloadPodScaleUp", deployment.VerifyDeploymentPodScaleUp},
			{"WorkloadPodScaleDown", deployment.VerifyDeploymentPodScaleDown},
			{"WorkloadPauseOrchestration", deployment.VerifyDeploymentPauseOrchestration},
		}

		for _, workloadTest := range workloadTests {
			p.Run(workloadTest.name, func() {
				for _, clusterID := range clusterIDs {
					clusterName, err := clusterExtensions.GetClusterNameByID(p.client, clusterID)
					require.NoError(p.T(), err)

					logrus.Infof("Running %s on cluster %s", workloadTest.name, clusterName)
					retries := 3
					for i := 0; i+1 < retries; i++ {
						err := workloadTest.validationFunc(p.client, clusterID)
						if err != nil {
							logrus.Info(err)
							logrus.Infof("Retry %v / %v", i+1, retries)
							continue
						}

						break
					}
				}
			})
		}
	}

	if p.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestOSValidationTestSuite(t *testing.T) {
	suite.Run(t, new(OSValidationTestSuite))
}
