package registries

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/registries"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TfpRegistriesTestSuite struct {
	suite.Suite
	client                     *rancher.Client
	session                    *session.Session
	rancherConfig              *rancher.Config
	terraformConfig            *config.TerraformConfig
	terratestConfig            *config.TerratestConfig
	standaloneTerraformOptions *terraform.Options
	terraformOptions           *terraform.Options
	adminUser                  *management.User
	authRegistry               string
	nonAuthRegistry            string
	globalRegistry             string
}

func (r *TfpRegistriesTestSuite) TearDownSuite() {
	keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath)
	cleanup.Cleanup(r.T(), r.standaloneTerraformOptions, keyPath)
}

func (r *TfpRegistriesTestSuite) SetupSuite() {
	r.terraformConfig = new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, r.terraformConfig)

	r.terratestConfig = new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, r.terratestConfig)

	keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath)
	standaloneTerraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.standaloneTerraformOptions = standaloneTerraformOptions

	authRegistry, nonAuthRegistry, globalRegistry, err := registries.CreateMainTF(r.T(), r.standaloneTerraformOptions, keyPath, r.terraformConfig, r.terratestConfig)
	require.NoError(r.T(), err)

	r.authRegistry = authRegistry
	r.nonAuthRegistry = nonAuthRegistry
	r.globalRegistry = globalRegistry
}

func (r *TfpRegistriesTestSuite) TfpSetupSuite(terratestConfig *config.TerratestConfig, terraformConfig *config.TerraformConfig) {
	testSession := session.NewSession()
	r.session = testSession

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	r.rancherConfig = rancherConfig

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	r.adminUser = adminUser

	userToken, err := token.GenerateUserToken(adminUser, r.rancherConfig.Host)
	require.NoError(r.T(), err)

	rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(rancherConfig.AdminToken, testSession)
	require.NoError(r.T(), err)

	r.client = client
	r.client.RancherConfig.AdminToken = rancherConfig.AdminToken

	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	terraformOptions := framework.Setup(r.T(), terraformConfig, terratestConfig, keyPath)
	r.terraformOptions = terraformOptions
}

func (r *TfpRegistriesTestSuite) TestTfpGlobalRegistry() {
	nodeRolesAll := config.AllRolesNodePool
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Global RKE1", "ec2_rke1", nodeRolesDedicated},
		{"Global RKE2", "ec2_rke2", nodeRolesDedicated},
		{"Global K3S", "ec2_k3s", []config.Nodepool{nodeRolesAll}},
	}

	for _, tt := range tests {
		terratestConfig := *r.terratestConfig
		terraformConfig := *r.terraformConfig
		terratestConfig.Nodepools = tt.nodeRoles

		terraformConfig.Module = tt.module
		terraformConfig.PrivateRegistries.SystemDefaultRegistry = r.globalRegistry
		terraformConfig.PrivateRegistries.URL = r.globalRegistry
		terraformConfig.PrivateRegistries.Password = ""
		terraformConfig.PrivateRegistries.Username = ""
		terraformConfig.StandaloneRegistry.Authenticated = false

		r.TfpSetupSuite(&terratestConfig, &terraformConfig)

		provisioning.GetK8sVersion(r.T(), r.client, &terratestConfig, &terraformConfig, configs.DefaultK8sVersion)

		tt.name = tt.name + " Kubernetes version: " + terratestConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(r.T(), r.client, r.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, r.terraformOptions, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], &terraformConfig)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (r *TfpRegistriesTestSuite) TestTfpAuthenticatedRegistry() {
	nodeRolesAll := config.AllRolesNodePool
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Auth RKE1", "ec2_rke1", nodeRolesDedicated},
		{"Auth RKE2", "ec2_rke2", nodeRolesDedicated},
		{"Auth K3S", "ec2_k3s", []config.Nodepool{nodeRolesAll}},
	}

	for _, tt := range tests {
		terratestConfig := *r.terratestConfig
		terraformConfig := *r.terraformConfig
		terratestConfig.Nodepools = tt.nodeRoles
		terraformConfig.Module = tt.module

		terraformConfig.PrivateRegistries.SystemDefaultRegistry = r.authRegistry
		terraformConfig.PrivateRegistries.URL = r.authRegistry
		terraformConfig.StandaloneRegistry.Authenticated = true

		r.TfpSetupSuite(&terratestConfig, &terraformConfig)

		provisioning.GetK8sVersion(r.T(), r.client, &terratestConfig, &terraformConfig, configs.DefaultK8sVersion)

		tt.name = tt.name + " Kubernetes version: " + terratestConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(r.T(), r.client, r.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, r.terraformOptions, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], &terraformConfig)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func (r *TfpRegistriesTestSuite) TestTfpNonAuthenticatedRegistry() {
	nodeRolesAll := config.AllRolesNodePool
	nodeRolesDedicated := []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}

	tests := []struct {
		name      string
		module    string
		nodeRoles []config.Nodepool
	}{
		{"Non Auth RKE1", "ec2_rke1", nodeRolesDedicated},
		{"Non Auth RKE2", "ec2_rke2", nodeRolesDedicated},
		{"Non Auth K3S", "ec2_k3s", []config.Nodepool{nodeRolesAll}},
	}

	for _, tt := range tests {
		terratestConfig := *r.terratestConfig
		terraformConfig := *r.terraformConfig
		terratestConfig.Nodepools = tt.nodeRoles

		terraformConfig.Module = tt.module
		terraformConfig.PrivateRegistries.SystemDefaultRegistry = r.nonAuthRegistry
		terraformConfig.PrivateRegistries.URL = r.nonAuthRegistry
		terraformConfig.PrivateRegistries.Password = ""
		terraformConfig.PrivateRegistries.Username = ""
		terraformConfig.StandaloneRegistry.Authenticated = false

		r.TfpSetupSuite(&terratestConfig, &terraformConfig)

		provisioning.GetK8sVersion(r.T(), r.client, &terratestConfig, &terraformConfig, configs.DefaultK8sVersion)

		tt.name = tt.name + " Kubernetes version: " + terratestConfig.KubernetesVersion
		testUser, testPassword, clusterName, poolName := configs.CreateTestCredentials()

		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			clusterIDs := provisioning.Provision(r.T(), r.client, r.rancherConfig, &terraformConfig, &terratestConfig, testUser, testPassword, clusterName, poolName, r.terraformOptions, nil)
			provisioning.VerifyClustersState(r.T(), r.client, clusterIDs)
			provisioning.VerifyRegistry(r.T(), r.client, clusterIDs[0], &terraformConfig)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpRegistriesTestSuite(t *testing.T) {
	suite.Run(t, new(TfpRegistriesTestSuite))
}
