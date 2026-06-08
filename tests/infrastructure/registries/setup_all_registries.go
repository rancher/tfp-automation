package registries

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	registry "github.com/rancher/tfp-automation/framework/set/resources/registries/createRegistry"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	authRegistryPublicDNS    = "auth_registry_public_dns"
	nonAuthRegistryPublicDNS = "non_auth_registry_public_dns"
	globalRegistryPublicDNS  = "global_registry_public_dns"
	ecrRegistryPublicDNS     = "ecr_registry_public_dns"

	authGlobalRegistryRoute53FQDN    = "auth_global_registry_route_53_fqdn"
	nonAuthGlobalRegistryRoute53FQDN = "nauth_global_registry_route_53_fqdn"

	authRegistry    = "auth_registry"
	nonAuthRegistry = "non_auth_registry"
	globalRegistry  = "global_registry"
	ecrRegistry     = "ecr_registry"

	terraformConst = "terraform"
)

// SetupAllRegistries is a function that creates all registries in a Rancher setup, either via CLI or web application
func SetupAllRegistries(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath, terratestConfig.PathToRepo, "all")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	terraformConfig.StandaloneRegistry.Enabled = true

	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{authRegistry, nonAuthRegistry, globalRegistry, ecrRegistry}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	require.NoError(t, err)

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating resources. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return err
	}

	authRegistryPublicDNS := terraform.Output(t, terraformOptions, authRegistryPublicDNS)
	nonAuthRegistryPublicDNS := terraform.Output(t, terraformOptions, nonAuthRegistryPublicDNS)
	globalRegistryPublicDNS := terraform.Output(t, terraformOptions, globalRegistryPublicDNS)
	ecrRegistryPublicDNS := terraform.Output(t, terraformOptions, ecrRegistryPublicDNS)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating non-authenticated registry...")
	file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, nonAuthRegistryPublicDNS, nonAuthRegistry, "", false)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating global registry...")
	file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, globalRegistryPublicDNS, globalRegistry, "", false)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating authenticated registry...")
	file, err = registry.CreateAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, authRegistryPublicDNS, authRegistry, authRegistryPublicDNS, false)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating ecr registry...")
	file, err = registry.CreateECRRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, ecrRegistryPublicDNS)
	require.NoError(t, err)

	terraform.InitAndApply(t, terraformOptions)

	return nil
}
