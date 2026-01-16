package rancher2

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
	vsphereDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/defaults/rke"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	admin                   = "admin"
	apiURL                  = "api_url"
	alias                   = "alias"
	allowUnverifiedSSL      = "allow_unverified_ssl"
	ec2                     = "ec2"
	globalRoleBinding       = "rancher2_global_role_binding"
	globalRoleID            = "global_role_id"
	insecure                = "insecure"
	name                    = "name"
	password                = "password"
	provider                = "provider"
	rancher2Const           = "rancher2"
	rancher2CustomUserToken = "rancher2_custom_user_token"
	rancherRKE              = "rancher/rke"
	rancherSource           = "source"
	rancherUser             = "rancher2_user"
	rc                      = "-rc"
	requiredProviders       = "required_providers"
	terraform               = "terraform"
	testPassword            = "password"
	tokenKey                = "token_key"
	ttl                     = "ttl"
	version                 = "version"
	vsphere                 = "vsphere"
	user                    = "user"
	userID                  = "user_id"
	username                = "username"
	providerEnvVar          = "RANCHER2_PROVIDER_VERSION"
	cloudProviderEnvVar     = "CLOUD_PROVIDER_VERSION"
	localProviderEnvVar     = "LOCALS_PROVIDER_VERSION"
	rkeEnvVar               = "RKE_PROVIDER_VERSION"
)

// SetProvidersAndUsersTF is a helper function that will set the general Terraform configurations in the main.tf file.
func SetProvidersAndUsersTF(rancherConfig *rancher.Config, testUser, testPassword string, authProvider bool,
	newFile *hclwrite.File, rootBody *hclwrite.Body, configMap []map[string]any, customModule bool) (*hclwrite.File, *hclwrite.Body) {
	createRequiredProviders(rootBody, configMap, customModule)
	createProvider(rancherConfig, rootBody, configMap, customModule)
	createProviderAlias(rancherConfig, rootBody)

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

	source, rancherProviderVersion, cloudProviderVersion, localProviderVersion, rkeProviderVersion := getRequiredProviderVersions(configMap)

	if rancherProviderVersion != "" {
		reqProvsBlockBody.SetAttributeValue(rancher2Const, cty.ObjectVal(map[string]cty.Value{
			rancherSource: cty.StringVal(source),
			version:       cty.StringVal(rancherProviderVersion),
		}))
	}

	if cloudProviderVersion != "" && terraformConfig.Provider == aws.Aws && customModule {
		reqProvsBlockBody.SetAttributeValue(aws.Aws, cty.ObjectVal(map[string]cty.Value{
			general.Source:  cty.StringVal(aws.AwsSource),
			general.Version: cty.StringVal(cloudProviderVersion),
		}))
	}

	if cloudProviderVersion != "" && terraformConfig.Provider == linode.Linode && customModule {
		reqProvsBlockBody.SetAttributeValue(linode.Linode, cty.ObjectVal(map[string]cty.Value{
			general.Source:  cty.StringVal(linode.LinodeSource),
			general.Version: cty.StringVal(cloudProviderVersion),
		}))
	}

	if cloudProviderVersion != "" && terraformConfig.Provider == vsphereDefaults.Vsphere && customModule {
		reqProvsBlockBody.SetAttributeValue(vsphereDefaults.Vsphere, cty.ObjectVal(map[string]cty.Value{
			general.Source:  cty.StringVal(vsphereDefaults.VsphereSource),
			general.Version: cty.StringVal(cloudProviderVersion),
		}))
	}

	if localProviderVersion != "" {
		reqProvsBlockBody.SetAttributeValue(general.Local, cty.ObjectVal(map[string]cty.Value{
			general.Source:  cty.StringVal(general.LocalSource),
			general.Version: cty.StringVal(localProviderVersion),
		}))
	}

	if rkeProviderVersion != "" {
		reqProvsBlockBody.SetAttributeValue(rke.RKE, cty.ObjectVal(map[string]cty.Value{
			general.Source:  cty.StringVal(rancherRKE),
			general.Version: cty.StringVal(rkeProviderVersion),
		}))
	}

	rootBody.AppendNewline()
}

// createProvider creates a provider block for the given rancher config.
func createProvider(rancherConfig *rancher.Config, rootBody *hclwrite.Body, configMap []map[string]any, customModule bool) {
	_, _, cloudProviderVersion, _, _ := getRequiredProviderVersions(configMap)

	terraformConfig := new(config.TerraformConfig)
	operations.LoadObjectFromMap(config.TerraformConfigurationFileKey, configMap[0], terraformConfig)

	if cloudProviderVersion != "" && terraformConfig.Provider == aws.Aws && customModule {
		awsProvBlock := rootBody.AppendNewBlock(general.Provider, []string{aws.Aws})
		awsProvBlockBody := awsProvBlock.Body()

		awsProvBlockBody.SetAttributeValue(aws.Region, cty.StringVal(terraformConfig.AWSConfig.Region))
		awsProvBlockBody.SetAttributeValue(aws.AccessKey, cty.StringVal(terraformConfig.AWSCredentials.AWSAccessKey))
		awsProvBlockBody.SetAttributeValue(aws.SecretKey, cty.StringVal(terraformConfig.AWSCredentials.AWSSecretKey))

		rootBody.AppendNewline()
		rootBody.AppendNewBlock(general.Provider, []string{general.Local})
		rootBody.AppendNewline()
	}

	if cloudProviderVersion != "" && terraformConfig.Provider == linode.Linode && customModule {
		linodeProvBlock := rootBody.AppendNewBlock(general.Provider, []string{linode.Linode})
		linodeProvBlockBody := linodeProvBlock.Body()

		linodeProvBlockBody.SetAttributeValue(general.Token, cty.StringVal(terraformConfig.LinodeCredentials.LinodeToken))

		rootBody.AppendNewline()
		rootBody.AppendNewBlock(general.Provider, []string{general.Local})
		rootBody.AppendNewline()
	}

	if cloudProviderVersion != "" && terraformConfig.Provider == vsphereDefaults.Vsphere && customModule {
		vsphereProvBlock := rootBody.AppendNewBlock(general.Provider, []string{vsphereDefaults.Vsphere})
		vsphereProvBlockBody := vsphereProvBlock.Body()

		vsphereProvBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.VsphereCredentials.Username))
		vsphereProvBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.VsphereCredentials.Password))
		vsphereProvBlockBody.SetAttributeValue(vsphereDefaults.VsphereServer, cty.StringVal(terraformConfig.VsphereCredentials.Vcenter))
		vsphereProvBlockBody.SetAttributeValue(allowUnverifiedSSL, cty.BoolVal(true))

		rootBody.AppendNewline()
		rootBody.AppendNewBlock(general.Provider, []string{general.Local})
		rootBody.AppendNewline()
	}

	rancher2ProvBlock := rootBody.AppendNewBlock(provider, []string{rancher2Const})
	rancher2ProvBlockBody := rancher2ProvBlock.Body()

	rancher2ProvBlockBody.SetAttributeValue(apiURL, cty.StringVal("https://"+rancherConfig.Host))
	rancher2ProvBlockBody.SetAttributeValue(tokenKey, cty.StringVal(rancherConfig.AdminToken))
	rancher2ProvBlockBody.SetAttributeValue(insecure, cty.BoolVal(*rancherConfig.Insecure))

	rootBody.AppendNewline()
}

// createProviderAlias creates a provider alias block for the standard user.
func createProviderAlias(rancherConfig *rancher.Config, rootBody *hclwrite.Body) {
	providerBlock := rootBody.AppendNewBlock(general.Provider, []string{rancher2Const})
	providerBlockBody := providerBlock.Body()

	providerBlockBody.SetAttributeValue(alias, cty.StringVal(general.AdminUser))
	providerBlockBody.SetAttributeValue(apiURL, cty.StringVal("https://"+rancherConfig.Host))

	adminUser := &management.User{
		Username: admin,
		Password: rancherConfig.AdminPassword,
	}

	adminToken, err := token.GenerateUserToken(adminUser, rancherConfig.Host)
	if err != nil {
		logrus.Fatalf("Failed to generate admin token: %v", err)
	}

	providerBlockBody.SetAttributeValue(tokenKey, cty.StringVal(adminToken.Token))
	providerBlockBody.SetAttributeValue(insecure, cty.BoolVal(true))

	rootBody.AppendNewline()
}

// Determines the required providers from the list of configs.
func getRequiredProviderVersions(configMap []map[string]any) (source, rancherProviderVersion, rkeProviderVersion, localProviderVersion,
	cloudProviderVersion string) {
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

		if strings.Contains(module, general.Import) && strings.Contains(module, clustertypes.RKE1) {
			rkeProviderVersion = os.Getenv(rkeEnvVar)
			if rkeProviderVersion == "" {
				logrus.Fatalf("Expected env var not set %s", rkeEnvVar)
			}
		}

		if strings.Contains(module, general.Custom) || strings.Contains(module, general.Import) || strings.Contains(module, general.Airgap) ||
			strings.Contains(module, ec2) {
			cloudProviderVersion = os.Getenv(cloudProviderEnvVar)
			if cloudProviderVersion == "" {
				logrus.Fatalf("Expected env var not set %s", cloudProviderEnvVar)
			}

			localProviderVersion = os.Getenv(localProviderEnvVar)
			if localProviderVersion == "" {
				logrus.Fatalf("Expected env var not set %s", localProviderEnvVar)
			}
		}
	}

	return source, rancherProviderVersion, cloudProviderVersion, localProviderVersion, rkeProviderVersion
}
