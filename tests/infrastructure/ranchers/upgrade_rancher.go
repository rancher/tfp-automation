package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/features"
	reg "github.com/rancher/tests/actions/registries"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	featureDefaults "github.com/rancher/tfp-automation/framework/set/defaults/features"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	"github.com/stretchr/testify/require"
)

// UpgradeAirgapRancher upgrades an existing airgapped Rancher server and returns the client, configuration, and Terraform options.
func UpgradeAirgapRancher(t *testing.T, client *rancher.Client, bastion, registry string, session *session.Session, cattleConfig map[string]any,
	tunnel *ssh.BastionSSHTunnel) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeAirgapRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, "", "", bastion, registry)
	require.NoError(t, err)

	tunnel.StopBastionSSHTunnel()

	session = session.NewSession()

	sshKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	require.NoError(t, err)

	tunnel, err = ssh.StartBastionSSHTunnel(bastion, terraformConfig.Standalone.OSUser, sshKey, "8443", standaloneConfig.RancherHostname, "443")
	require.NoError(t, err)

	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.AirgapKeyPath)
	client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	if standaloneConfig.UpgradedRancherTagVersion != head {
		_, airgapKeyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
		terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, airgapKeyPath)

		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.UpgradedRancherTagVersion, airgapKeyPath, terraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	usesRegistryPrefix, err := reg.CheckAllClusterPodsForRegistryPrefix(client, local, registry)
	require.NoError(t, err)

	if !usesRegistryPrefix {
		t.Fatalf("ERROR: not all of the local cluster pods are using the private registry")
	}

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}

// UpgradeDualStackRancher upgrades an existing dual-stack Rancher server and returns the client, configuration, and Terraform options.
func UpgradeDualStackRancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeDualStackRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)

	session = session.NewSession()

	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.DualStackKeyPath)
	client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	if standaloneConfig.UpgradedRancherTagVersion != head {
		_, dualStackKeyPath := rancher2.SetKeyPath(keypath.DualStackKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
		terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, dualStackKeyPath)

		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.UpgradedRancherTagVersion, dualStackKeyPath, terraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}

// UpgradeIPv6Rancher upgrades an existing IPv6 Rancher server and returns the client, configuration, and Terraform options.
func UpgradeIPv6Rancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeIPv6Rancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)

	session = session.NewSession()

	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.IPv6KeyPath)
	client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	if standaloneConfig.UpgradedRancherTagVersion != head {
		_, ipv6KeyPath := rancher2.SetKeyPath(keypath.IPv6KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
		terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, ipv6KeyPath)

		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.UpgradedRancherTagVersion, ipv6KeyPath, terraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}

// UpgradeProxyRancher upgrades an existing proxy Rancher server and returns the client, configuration, and Terraform options.
func UpgradeProxyRancher(t *testing.T, client *rancher.Client, proxyPrivateIP, proxyBastion string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client,
	map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeProxyRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, proxyPrivateIP, proxyBastion, "", "")
	require.NoError(t, err)

	session = session.NewSession()

	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.ProxyKeyPath)
	client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	if standaloneConfig.UpgradedRancherTagVersion != head {
		_, proxyKeyPath := rancher2.SetKeyPath(keypath.ProxyKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
		terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, proxyKeyPath)

		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.UpgradedRancherTagVersion, proxyKeyPath, terraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}

// UpgradeRancher upgrades an existing Rancher server and returns the client, configuration, and Terraform options.
func UpgradeRancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)

	session = session.NewSession()

	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.SanityKeyPath)
	client, err = PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	if standaloneConfig.UpgradedRancherTagVersion != head {
		_, sanityKeyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
		terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, sanityKeyPath)

		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.UpgradedRancherTagVersion, sanityKeyPath, terraformOptions)
	}

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.UpgradedTurtles != "" {
		switch standaloneConfig.FeatureFlags.UpgradedTurtles {
		case featureDefaults.ToggledOff:
			features.UpdateFeatureFlag(client, featureDefaults.Turtles, false)
		case featureDefaults.ToggledOn:
			features.UpdateFeatureFlag(client, featureDefaults.Turtles, true)
		}
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}
