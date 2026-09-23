package authprovider

import (
	"fmt"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/authproviders"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ad"
	"github.com/rancher/tfp-automation/framework/set/authproviders/keycloakSAML"
	"github.com/rancher/tfp-automation/framework/set/authproviders/ldap"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/stretchr/testify/require"
)

// Enable is a function that will run terraform apply to enable the auth provider.
func Enable(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, terraformOptions *terraform.Options) {
	apply(t, rancherConfig, terraformConfig, terratestConfig, terraformOptions, true)
}

// Disable is a function that will run terraform apply to disable the auth provider in place.
func Disable(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, terraformOptions *terraform.Options) {
	apply(t, rancherConfig, terraformConfig, terratestConfig, terraformOptions, false)
}

// Destroy is a function that will run terraform destroy to remove the auth provider resource.
func Destroy(t *testing.T, terraformOptions *terraform.Options) {
	terraform.Destroy(t, terraformOptions)
}

// apply is a function that will write the auth provider main.tf and apply it with the given enabled state.
func apply(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, terraformOptions *terraform.Options, enabled bool) {
	newFile, rootBody, file := resources.InitializeMainTF(terratestConfig)
	require.NotNil(t, file)

	defer file.Close()

	newFile, rootBody = resources.SetProvidersAndUsersTF(rancherConfig, true, newFile, rootBody, terraformConfig, false)

	err := setAuthProvider(rancherConfig, terraformConfig, enabled, newFile, rootBody, file)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)
}

// setAuthProvider is a function that will set the auth provider resource with the given enabled state.
func setAuthProvider(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, enabled bool,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) error {
	switch terraformConfig.AuthProvider {
	case authproviders.AD:
		terraformConfig.ADConfig.Enabled = &enabled
		return ad.SetAD(terraformConfig, newFile, rootBody, file)
	case authproviders.KeycloakSAML:
		terraformConfig.KeycloakSAMLConfig.Enabled = &enabled
		return keycloakSAML.SetKeycloakSAML(rancherConfig, terraformConfig, newFile, rootBody, file)
	case authproviders.OpenLDAP:
		terraformConfig.OpenLDAPConfig.Enabled = &enabled
		return ldap.SetOpenLDAP(terraformConfig, newFile, rootBody, file)
	default:
		return fmt.Errorf("unsupported auth provider: %s", terraformConfig.AuthProvider)
	}
}
