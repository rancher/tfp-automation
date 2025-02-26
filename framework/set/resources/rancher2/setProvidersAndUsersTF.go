package rancher2

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	apiURL              = "api_url"
	globalRoleBinding   = "rancher2_global_role_binding"
	globalRoleID        = "global_role_id"
	insecure            = "insecure"
	name                = "name"
	provider            = "provider"
	rancher2            = "rancher2"
	rancherRKE          = "rancher/rke"
	rancherSource       = "source"
	rancherUser         = "rancher2_user"
	rc                  = "-rc"
	requiredProviders   = "required_providers"
	terraform           = "terraform"
	testPassword        = "password"
	tokenKey            = "token_key"
	version             = "version"
	user                = "user"
	userID              = "user_id"
	username            = "username"
	providerEnvVar      = "RANCHER2_PROVIDER_VERSION"
	awsProviderEnvVar   = "AWS_PROVIDER_VERSION"
	localProviderEnvVar = "LOCALS_PROVIDER_VERSION"
	rkeEnvVar           = "RKE_PROVIDER_VERSION"
)

// SetProvidersAndUsersTF is a helper function that will set the general Terraform configurations in the main.tf file.
func SetProvidersAndUsersTF(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, testUser, testPassword string, authProvider bool, configMap []map[string]any) (*hclwrite.File, *hclwrite.Body) {
	providerVersion, awsProviderVersion, localProviderVersion, source, rkeProviderVersion := getProviderVersions(terraformConfig, configMap)

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	createRequiredProviders(rootBody, terraformConfig, awsProviderVersion, localProviderVersion, providerVersion, source, rkeProviderVersion, configMap)

	rootBody.AppendNewline()

	createProvider(rootBody, rancherConfig, terraformConfig, configMap)

	createUser(rootBody, testUser, testPassword)

	if !authProvider {
		createGlobalRoleBinding(rootBody, testUser, userID)
	}

	return newFile, rootBody
}

// getProviderVersions returns the versions for the providers based on environment variables.
func getProviderVersions(terraformConfig *config.TerraformConfig, configMap []map[string]any) (string, string, string, string, string) {
	terraform := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

	customModule := containsCustomModule(configMap)

	providerVersion := os.Getenv(providerEnvVar)
	if providerVersion == "" {
		logrus.Fatalf("Expected env var not set %s", providerEnvVar)
	}

	var awsProviderVersion, localProviderVersion, rkeProviderVersion string

	if strings.Contains(terraform.Module, defaults.Custom) || strings.Contains(terraform.Module, defaults.Airgap) ||
		strings.Contains(terraform.Module, defaults.Import) || customModule {
		awsProviderVersion = os.Getenv(awsProviderEnvVar)
		if awsProviderVersion == "" {
			logrus.Fatalf("Expected env var not set %s", awsProviderEnvVar)
		}

		localProviderVersion = os.Getenv(localProviderEnvVar)
		if providerVersion == "" {
			logrus.Fatalf("Expected env var not set %s", localProviderEnvVar)
		}
	}

	source := "rancher/rancher2"
	if strings.Contains(providerVersion, rc) {
		source = "terraform.local/local/rancher2"
	}

	var importCluster bool
	for _, cattleConfig := range configMap {
		terraform := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, terraform)
		if terraform.Module == modules.ImportEC2RKE1 {
			importCluster = true
		}
	}

	if importCluster || terraformConfig.Module == modules.ImportEC2RKE1 {
		rkeProviderVersion = os.Getenv(rkeEnvVar)
		if rkeProviderVersion == "" {
			logrus.Fatalf("Expected env var not set %s", rkeEnvVar)
		}
	}

	return providerVersion, awsProviderVersion, localProviderVersion, source, rkeProviderVersion
}

// createRequiredProviders creates the required_providers block.
func createRequiredProviders(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, awsProviderVersion, localProviderVersion,
	providerVersion, source, rkeProviderVersion string, configMap []map[string]any) {
	tfConfig := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], tfConfig)

	tfBlock := rootBody.AppendNewBlock(terraform, nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	customModule := containsCustomModule(configMap)

	if strings.Contains(tfConfig.Module, defaults.Airgap) || strings.Contains(tfConfig.Module, defaults.Import) ||
		customModule {
		reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(defaults.AwsSource),
			defaults.Version: cty.StringVal(awsProviderVersion),
		}))

		reqProvsBlockBody.SetAttributeValue(defaults.Local, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(defaults.LocalSource),
			defaults.Version: cty.StringVal(localProviderVersion),
		}))
	}

	reqProvsBlockBody.SetAttributeValue(rancher2, cty.ObjectVal(map[string]cty.Value{
		rancherSource: cty.StringVal(source),
		version:       cty.StringVal(providerVersion),
	}))

	var importCluster bool
	for _, cattleConfig := range configMap {
		terraform := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, terraform)

		if terraform.Module == modules.ImportEC2RKE1 {
			importCluster = true
		}
	}

	if importCluster || terraformConfig.Module == modules.ImportEC2RKE1 {
		reqProvsBlockBody.SetAttributeValue(defaults.RKE, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(rancherRKE),
			defaults.Version: cty.StringVal(rkeProviderVersion),
		}))
	}
}

// createProvider creates a provider block for the given rancher config.
func createProvider(rootBody *hclwrite.Body, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, configMap []map[string]any) {
	terraform := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraform)

	customModule := containsCustomModule(configMap)

	if strings.Contains(terraform.Module, defaults.Airgap) || strings.Contains(terraform.Module, defaults.Import) || customModule {
		awsProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Aws})
		awsProvBlockBody := awsProvBlock.Body()

		awsProvBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
		awsProvBlockBody.SetAttributeValue(defaults.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
		awsProvBlockBody.SetAttributeValue(defaults.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))

		rootBody.AppendNewline()
		rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Local})
		rootBody.AppendNewline()
	}

	rancher2ProvBlock := rootBody.AppendNewBlock(provider, []string{rancher2})
	rancher2ProvBlockBody := rancher2ProvBlock.Body()

	rancher2ProvBlockBody.SetAttributeValue(apiURL, cty.StringVal("https://"+rancherConfig.Host))
	rancher2ProvBlockBody.SetAttributeValue(tokenKey, cty.StringVal(rancherConfig.AdminToken))
	rancher2ProvBlockBody.SetAttributeValue(insecure, cty.BoolVal(*rancherConfig.Insecure))

	rootBody.AppendNewline()
}

// createUser creates the user block for a new user.
func createUser(rootBody *hclwrite.Body, testUser, testpassword string) {
	userBlock := rootBody.AppendNewBlock(defaults.Resource, []string{rancherUser, rancherUser})
	userBlockBody := userBlock.Body()

	userBlockBody.SetAttributeValue(name, cty.StringVal(testUser))
	userBlockBody.SetAttributeValue(username, cty.StringVal(testUser))
	userBlockBody.SetAttributeValue(testPassword, cty.StringVal(testpassword))
	userBlockBody.SetAttributeValue(defaults.Enabled, cty.BoolVal(true))

	rootBody.AppendNewline()
}

// createGlobalRoleBinding creates a global role binding block for the given user.
func createGlobalRoleBinding(rootBody *hclwrite.Body, testUser string, userID string) {
	globalRoleBindingBlock := rootBody.AppendNewBlock(defaults.Resource, []string{globalRoleBinding, globalRoleBinding})
	globalRoleBindingBlockBody := globalRoleBindingBlock.Body()

	globalRoleBindingBlockBody.SetAttributeValue(name, cty.StringVal(testUser))
	globalRoleBindingBlockBody.SetAttributeValue(globalRoleID, cty.StringVal(user))

	standardUser := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancherUser + "." + rancherUser + ".id")},
	}

	globalRoleBindingBlockBody.SetAttributeRaw(userID, standardUser)
}

// determine if config within configMap contains a custom module
func containsCustomModule(configMap []map[string]any) bool {
	customModule := false

	for _, cattleConfig := range configMap {
		tfConfig := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, tfConfig)
		module := tfConfig.Module

		if strings.Contains(module, defaults.Custom) {
			customModule = true
		}
	}
	return customModule
}
