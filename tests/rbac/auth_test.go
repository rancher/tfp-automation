package rbac

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	rb "github.com/rancher/tfp-automation/tests/extensions/rbac"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthConfigTestSuite struct {
	suite.Suite
	client           *rancher.Client
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	clusterConfig    *config.TerratestConfig
	terraformOptions *terraform.Options
}

func (r *AuthConfigTestSuite) SetupSuite() {
	testSession := session.NewSession()
	r.session = testSession

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	r.terraformConfig = terraformConfig

	terraformOptions := framework.Setup(r.T())
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
			defer cleanup.ConfigCleanup(r.T(), r.terraformOptions)

			rb.AuthConfig(r.T(), &authConfig, r.terraformOptions)
		})
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
			defer cleanup.ConfigCleanup(r.T(), r.terraformOptions)

			rb.AuthConfig(r.T(), r.terraformConfig, r.terraformOptions)
		})
	}
}

func TestTfpAuthConfigTestSuite(t *testing.T) {
	suite.Run(t, new(AuthConfigTestSuite))
}
