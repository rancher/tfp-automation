package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/featureflags"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/dualstack"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6"
	proxy "github.com/rancher/tfp-automation/framework/set/resources/proxy"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/registries"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
)

// SetupAirgapRancher sets up an airgapped Rancher server and returns the client, configuration, and Terraform options.
func SetupAirgapRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	registry, bastion, err := airgap.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.AirgapInternalFQDN, keyPath, true, false)
	require.NoError(t, err)

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, registry, bastion, standaloneTerraformOptions, terraformOptions, cattleConfig
}

// SetupDualStackRancher sets up an dual-stack Rancher server and returns the client, configuration, and Terraform options.
func SetupDualStackRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	serverNodeOne, err := dualstack.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != "head" {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, serverNodeOne, standaloneTerraformOptions, terraformOptions, cattleConfig
}

// SetupIPv6Rancher sets up an IPv6-only Rancher server and returns the client, configuration, and Terraform options.
func SetupIPv6Rancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	bastionNode, err := ipv6.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != "head" {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, bastionNode, standaloneTerraformOptions, terraformOptions, cattleConfig
}

// SetupProxyRancher sets up a proxy Rancher server and returns the client, configuration, and Terraform options.
func SetupProxyRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, string,
	*terraform.Options, *terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	proxyBastion, proxyPrivateIP, err := proxy.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != "head" {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, proxyBastion, proxyPrivateIP, standaloneTerraformOptions, terraformOptions, cattleConfig
}

// SetupRancher sets up a Rancher server and returns the client, configuration, and Terraform options.
func SetupRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	serverNodeOne, err := sanity.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != "head" {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
		switch standaloneConfig.FeatureFlags.Turtles {
		case defaults.ToggledOff:
			featureflags.UpdateFeatureFlag(client, defaults.Turtles, false)
		case defaults.ToggledOn:
			featureflags.UpdateFeatureFlag(client, defaults.Turtles, true)
		}
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, serverNodeOne, standaloneTerraformOptions, terraformOptions, cattleConfig
}

// SetupRegistryRancher sets up a registry-enabled Rancher server and returns the client, configuration, and Terraform options.
func SetupRegistryRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, string, string, *terraform.Options, *terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	authRegistry, nonAuthRegistry, globalRegistry, err := registries.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != "head" {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, authRegistry, nonAuthRegistry, globalRegistry, standaloneTerraformOptions, terraformOptions, cattleConfig
}
