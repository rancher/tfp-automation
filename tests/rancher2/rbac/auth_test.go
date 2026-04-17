//go:build validation

package rbac

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/qase"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	tfpQase "github.com/rancher/tfp-automation/pipeline/qase"
	"github.com/rancher/tfp-automation/pipeline/qase/results"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/rancher/tfp-automation/tests/infrastructure/ranchers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthConfigTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (r *AuthConfigTestSuite) SetupSuite() {
	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig, _ = config.LoadTFPConfigs(r.cattleConfig)

	testSession := session.NewSession()
	r.session = testSession

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)

	r.terraformOptions = terraformOptions

	client, err := ranchers.PostRancherSetup(r.T(), r.terraformOptions, r.rancherConfig, r.session, r.rancherConfig.Host, keyPath, false)
	require.NoError(r.T(), err)

	r.client = client
}

func (r *AuthConfigTestSuite) TestTfpAuthConfig() {
	tests := []struct {
		name         string
		authProvider string
	}{
		{"Azure_AD", authproviders.AzureAD},
		{"GitHub", authproviders.GitHub},
		{"Okta", authproviders.Okta},
		{"OpenLDAP", authproviders.OpenLDAP},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		rancher, terraform, _, _ := config.LoadTFPConfigs(r.cattleConfig)
		rancher.AdminToken = r.client.RancherConfig.AdminToken
		terraform.AuthProvider = tt.authProvider

		terraform = provisioning.UniquifyTerraform(terraform)

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			rbac.AuthConfig(r.T(), rancher, terraform, r.terraformOptions, testUser, testPassword, []map[string]any{r.cattleConfig}, newFile, rootBody, file)
		})

		params := tfpQase.GetProvisioningSchemaParams(r.terraformConfig, r.terratestConfig)
		err := qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func (r *AuthConfigTestSuite) TestTfpAuthConfigDynamicInput() {
	if r.terraformConfig.AuthProvider == "" {
		r.T().Skip("No auth provider specified")
	}

	tests := []struct {
		name string
	}{
		{r.terraformConfig.AuthProvider},
	}

	testUser, testPassword := configs.CreateTestCredentials()

	for _, tt := range tests {
		newFile, rootBody, file := rancher2.InitializeMainTF(r.terratestConfig)
		defer file.Close()

		rancher, terraform, _, _ := config.LoadTFPConfigs(r.cattleConfig)
		rancher.AdminToken = r.client.RancherConfig.AdminToken
		terraform.AuthProvider = r.terraformConfig.AuthProvider

		terraform = provisioning.UniquifyTerraform(terraform)

		r.Run((tt.name), func() {
			_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.terratestConfig.PathToRepo, "")
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			rbac.AuthConfig(r.T(), rancher, terraform, r.terraformOptions, testUser, testPassword, []map[string]any{r.cattleConfig}, newFile, rootBody, file)
		})

		params := tfpQase.GetProvisioningSchemaParams(r.terraformConfig, r.terratestConfig)
		err := qase.UpdateSchemaParameters(tt.name, params)
		if err != nil {
			logrus.Warningf("Failed to upload schema parameters %s", err)
		}
	}

	if r.terratestConfig.LocalQaseReporting {
		results.ReportTest(r.terratestConfig)
	}
}

func TestTfpAuthConfigTestSuite(t *testing.T) {
	suite.Run(t, new(AuthConfigTestSuite))
}
