package rancher2

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	apiURL              = "api_url"
	ec2                 = "ec2"
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
func SetProvidersAndUsersTF(testUser, testPassword string, authProvider bool, newFile *hclwrite.File, rootBody *hclwrite.Body,
	configMap []map[string]any, customModule bool) (*hclwrite.File, *hclwrite.Body) {
	createRequiredProviders(rootBody, configMap, customModule)

	rootBody.AppendNewline()

	createProvider(rootBody, configMap, customModule)

	createUser(rootBody, testUser, testPassword)

	if !authProvider {
		createGlobalRoleBinding(rootBody, testUser, userID)
	}

	return newFile, rootBody
}

// createRequiredProviders creates the required_providers block.
func createRequiredProviders(rootBody *hclwrite.Body, configMap []map[string]any, customModule bool) {
	tfBlock := rootBody.AppendNewBlock(terraform, nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	terraformConfig := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraformConfig)

	source, rancherProviderVersion, awsProviderVersion, linodeProviderVersion, localProviderVersion, rkeProviderVersion := getRequiredProviderVersions(configMap)

	if rancherProviderVersion != "" {
		reqProvsBlockBody.SetAttributeValue(rancher2, cty.ObjectVal(map[string]cty.Value{
			rancherSource: cty.StringVal(source),
			version:       cty.StringVal(rancherProviderVersion),
		}))
	}

	if awsProviderVersion != "" && terraformConfig.Provider == defaults.Aws && customModule {
		reqProvsBlockBody.SetAttributeValue(defaults.Aws, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(defaults.AwsSource),
			defaults.Version: cty.StringVal(awsProviderVersion),
		}))
	}

	if linodeProviderVersion != "" && terraformConfig.Provider == defaults.Linode && customModule {
		reqProvsBlockBody.SetAttributeValue(defaults.Linode, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(defaults.LinodeSource),
			defaults.Version: cty.StringVal(linodeProviderVersion),
		}))
	}

	if localProviderVersion != "" {
		reqProvsBlockBody.SetAttributeValue(defaults.Local, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(defaults.LocalSource),
			defaults.Version: cty.StringVal(localProviderVersion),
		}))
	}

	if rkeProviderVersion != "" {
		reqProvsBlockBody.SetAttributeValue(defaults.RKE, cty.ObjectVal(map[string]cty.Value{
			defaults.Source:  cty.StringVal(rancherRKE),
			defaults.Version: cty.StringVal(rkeProviderVersion),
		}))
	}
}

// createProvider creates a provider block for the given rancher config.
func createProvider(rootBody *hclwrite.Body, configMap []map[string]any, customModule bool) {
	_, _, awsProviderVersion, linodeProviderVersion, _, _ := getRequiredProviderVersions(configMap)

	terraformConfig := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraformConfig)

	rancherConfig := new(rancher.Config)
	operations.LoadObjectFromMap(configs.Rancher, configMap[0], rancherConfig)

	if awsProviderVersion != "" && terraformConfig.Provider == defaults.Aws && customModule {
		awsProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Aws})
		awsProvBlockBody := awsProvBlock.Body()

		awsProvBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
		awsProvBlockBody.SetAttributeValue(defaults.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
		awsProvBlockBody.SetAttributeValue(defaults.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))

		rootBody.AppendNewline()
		rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Local})
		rootBody.AppendNewline()
	}

	if linodeProviderVersion != "" && terraformConfig.Provider == defaults.Linode && customModule {
		linodeProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Linode})
		linodeProvBlockBody := linodeProvBlock.Body()

		linodeProvBlockBody.SetAttributeValue(defaults.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))

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

// Determines the required providers from the list of configs.
func getRequiredProviderVersions(configMap []map[string]any) (source, rancherProviderVersion, rkeProviderVersion, localProviderVersion,
	awsProviderVersion, linodeProviderVersion string) {
	for _, cattleConfig := range configMap {
		terraformConfig := new(config.TerraformConfig)
		operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, cattleConfig, terraformConfig)
		module := terraformConfig.Module

		rancherProviderVersion = os.Getenv(providerEnvVar)
		if rancherProviderVersion == "" {
			logrus.Fatalf("Expected env var not set %s", providerEnvVar)
		}

		source = "rancher/rancher2"
		if strings.Contains(rancherProviderVersion, rc) {
			source = "terraform.local/local/rancher2"
		}

		if module == modules.ImportEC2RKE1 {
			rkeProviderVersion = os.Getenv(rkeEnvVar)
			if rkeProviderVersion == "" {
				logrus.Fatalf("Expected env var not set %s", rkeEnvVar)
			}
		}

		if strings.Contains(module, defaults.Custom) || strings.Contains(module, defaults.Import) || strings.Contains(module, defaults.Airgap) ||
			strings.Contains(module, ec2) {
			awsProviderVersion = os.Getenv(awsProviderEnvVar)
			if awsProviderVersion == "" {
				logrus.Fatalf("Expected env var not set %s", awsProviderEnvVar)
			}

			localProviderVersion = os.Getenv(localProviderEnvVar)
			if localProviderVersion == "" {
				logrus.Fatalf("Expected env var not set %s", localProviderEnvVar)
			}
		}
	}
	return source, rancherProviderVersion, awsProviderVersion, linodeProviderVersion, localProviderVersion, rkeProviderVersion
}
