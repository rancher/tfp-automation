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

func (o *OSValidationTestSuite) SetupSuite() {
	testSession := session.NewSession()
	o.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(o.T(), err)

	o.client = client

	o.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

	o.cattleConfig, err = config.LoadPackageDefaults(o.cattleConfig, "")
	require.NoError(o.T(), err)

	modulePermutation, err := permutationsdata.CreateModulePermutation(o.cattleConfig)
	require.NoError(o.T(), err)

	modulePermutation.KeyPathValueRelationships, err = permutationsdata.CreateAMIRelationships(o.cattleConfig)
	require.NoError(o.T(), err)

	k8sRelationships, err := permutationsdata.CreateK8sRelationships(o.cattleConfig)
	require.NoError(o.T(), err)

	modulePermutation.KeyPathValueRelationships = append(modulePermutation.KeyPathValueRelationships, k8sRelationships...)

	cniPermutation, err := permutationsdata.CreateCNIPermutation(o.cattleConfig)
	require.NoError(o.T(), err)

	permutedConfigs, err := permutations.Permute([]permutations.Permutation{*modulePermutation, *cniPermutation}, o.cattleConfig)
	require.NoError(o.T(), err)

	o.permutedConfigs, err = provisioning.UniquifyTerraform(permutedConfigs)
	require.NoError(o.T(), err)

	_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(o.permutedConfigs[0])

	o.awsCredentials = cloudcredentials.AmazonEC2CredentialConfig{
		AccessKey:     terraformConfig.AWSCredentials.AWSAccessKey,
		SecretKey:     terraformConfig.AWSCredentials.AWSSecretKey,
		DefaultRegion: terraformConfig.AWSConfig.Region,
	}

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, o.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(o.T(), terraformConfig, terratestConfig, keyPath)
	o.terraformOptions = terraformOptions
}

func (o *OSValidationTestSuite) TestDynamicOSValidation() {
	//Batch configs so that not all AMIs run in parallel
	configBatches := map[string][]map[string]any{}
	for _, cattleConfig := range o.permutedConfigs {
		_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)
		configBatches[terraformConfig.AWSConfig.AMI] = append(configBatches[terraformConfig.AWSConfig.AMI], cattleConfig)
	}

	newFile, rootBody, file := rancher2.InitializeMainTF(o.terratestConfig)
	defer file.Close()

	customClusterNames := []string{}

	for ami, batch := range configBatches {
		_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, o.terratestConfig.PathToRepo, "")
		defer cleanup.Cleanup(o.T(), o.terraformOptions, keyPath)
		testUser, testPassword := configs.CreateTestCredentials()

		var clusterIDs []string

		amiInfo, err := ec2.GetAMI(o.client, &o.awsCredentials, ami)
		require.NoError(o.T(), err)

		testName := "Parallel_Provisioning_" + *amiInfo.Images[0].Name
		o.Run(testName, func() {
			for _, cattleConfig := range batch {
				cattleConfig, err = operations.ReplaceValue([]string{"terraform", "awsConfig", "awsVolumeType"}, *amiInfo.Images[0].BlockDeviceMappings[0].Ebs.VolumeType, cattleConfig)
				require.NoError(o.T(), err)

				_, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

				logrus.Infof("Provisioning Cluster Type: %s, "+"K8s Version: %s, "+"CNI: %s", terraformConfig.Module, terratestConfig.KubernetesVersion, terraformConfig.CNI)
			}

			clusterIDs, _ = provisioning.Provision(o.T(), o.client, o.rancherConfig, o.terraformConfig, o.terratestConfig, testUser, testPassword, o.terraformOptions, batch, newFile, rootBody, file, false, false, true, customClusterNames)
			time.Sleep(2 * time.Minute)
			provisioning.VerifyClustersState(o.T(), o.client, clusterIDs)
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
			o.Run(workloadTest.name, func() {
				for _, clusterID := range clusterIDs {
					clusterName, err := clusterExtensions.GetClusterNameByID(o.client, clusterID)
					require.NoError(o.T(), err)

					logrus.Infof("Running %s on cluster %s", workloadTest.name, clusterName)
					retries := 3
					for i := 0; i+1 < retries; i++ {
						err := workloadTest.validationFunc(o.client, clusterID)
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

	if o.terratestConfig.LocalQaseReporting {
		qase.ReportTest(o.terratestConfig)
	}
}

func TestOSValidationTestSuite(t *testing.T) {
	suite.Run(t, new(OSValidationTestSuite))
}
