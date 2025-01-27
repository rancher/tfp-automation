package rbac

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	qase "github.com/rancher/tfp-automation/pipeline/qase/results"
	rb "github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthConfigTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	rancherConfig    *rancher.Config
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (r *AuthConfigTestSuite) SetupSuite() {
	testSession := session.NewSession()
	r.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	r.rancherConfig = rancherConfig

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	r.terraformConfig = terraformConfig

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	r.terratestConfig = terratestConfig

	keyPath := rancher2.SetKeyPath()
	terraformOptions := framework.Setup(r.T(), r.terraformConfig, r.terratestConfig, keyPath)
	r.terraformOptions = terraformOptions
}

func (r *AuthConfigTestSuite) TestTfpAuthConfig() {
	tests := []struct {
		name         string
		authProvider string
	}{
		{"Azure AD", authproviders.AzureAD},
		{"GitHub", authproviders.GitHub},
		{"Okta", authproviders.Okta},
		{"OpenLDAP", authproviders.OpenLDAP},
	}

	for _, tt := range tests {
		authConfig := *r.terraformConfig
		authConfig.AuthProvider = tt.authProvider
		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath()
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			rb.AuthConfig(r.T(), &authConfig, r.terraformOptions)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
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

	for _, tt := range tests {
		r.Run((tt.name), func() {
			keyPath := rancher2.SetKeyPath()
			defer cleanup.Cleanup(r.T(), r.terraformOptions, keyPath)

			rb.AuthConfig(r.T(), r.terraformConfig, r.terraformOptions)
		})
	}

	if r.terratestConfig.LocalQaseReporting {
		qase.ReportTest()
	}
}

func TestTfpAuthConfigTestSuite(t *testing.T) {
	suite.Run(t, new(AuthConfigTestSuite))
}
