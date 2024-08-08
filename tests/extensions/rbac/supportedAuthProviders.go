package rbac

import (
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
)

// SupportedAuthProviders is a function that will check if the user-inputted auth provider is supported.
func SupportedAuthProviders(terraformConfig *config.TerraformConfig, terraformOptions *terraform.Options) bool {
	authProvider := terraformConfig.AuthProvider
	supportedAuthProviders := []string{
		authproviders.AD,
		authproviders.AzureAD,
		authproviders.GitHub,
		authproviders.Okta,
		authproviders.OpenLDAP,
	}

	for _, supportedAuthProvider := range supportedAuthProviders {
		if authProvider == supportedAuthProvider {
			return true
		}
	}

	return false
}
