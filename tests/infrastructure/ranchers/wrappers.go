package ranchers

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	ranchersetup "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup"
	setupairgap "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/airgap"
	setupdualstack "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/dualstack"
	setuphosted "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/hosted"
	setupipv6 "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/ipv6"
	setupproxy "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/proxy"
	setupregistry "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/registry"
	setupstandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup/standard"
	upgradebase "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade"
	upgradeairgap "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/airgap"
	upgradedualstack "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/dualstack"
	upgradeipv6 "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/ipv6"
	upgradeproxy "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/proxy"
	upgradestandard "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/upgrade/standard"
)

func CreateRancher(t *testing.T, provider string) error {
	return setupstandard.CreateRancher(t, provider)
}

func CreateAirgapRancher(t *testing.T, provider string) error {
	return setupairgap.CreateAirgapRancher(t, provider)
}

func CreateDualStackRancher(t *testing.T, provider string) error {
	return setupdualstack.CreateDualStackRancher(t, provider)
}

func CreateHostedClusterRancher(t *testing.T, provider string) error {
	return setuphosted.CreateHostedClusterRancher(t, provider)
}

func CreateIPv6Rancher(t *testing.T, provider string) error {
	return setupipv6.CreateIPv6Rancher(t, provider)
}

func CreateProxyRancher(t *testing.T, provider string) error {
	return setupproxy.CreateProxyRancher(t, provider)
}

func CreateRegistryRancher(t *testing.T, provider string) error {
	return setupregistry.CreateRegistryRancher(t, provider)
}

func SetupAirgapRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, string, *terraform.Options,
	*terraform.Options, map[string]any, *ssh.BastionSSHTunnel) {
	return setupairgap.SetupAirgapRancher(t, session, moduleKeyPath)
}

func SetupDualStackRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	return setupdualstack.SetupDualStackRancher(t, session, moduleKeyPath)
}

func SetupIPv6Rancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	return setupipv6.SetupIPv6Rancher(t, session, moduleKeyPath)
}

func SetupProxyRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, string,
	*terraform.Options, *terraform.Options, map[string]any) {
	return setupproxy.SetupProxyRancher(t, session, moduleKeyPath)
}

func SetupRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	return setupstandard.SetupRancher(t, session, moduleKeyPath)
}

func SetupRegistryRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, string, string,
	*terraform.Options, *terraform.Options, map[string]any) {
	return setupregistry.SetupRegistryRancher(t, session, moduleKeyPath)
}

func UpgradingRancher(t *testing.T, provider string) error {
	return upgradestandard.UpgradingRancher(t, provider)
}

func UpgradingAirgapRancher(t *testing.T, provider string) error {
	return upgradeairgap.UpgradingAirgapRancher(t, provider)
}

func UpgradingDualStackRancher(t *testing.T, provider string) error {
	return upgradedualstack.UpgradingDualStackRancher(t, provider)
}

func UpgradingIPv6Rancher(t *testing.T, provider string) error {
	return upgradeipv6.UpgradingIPv6Rancher(t, provider)
}

func UpgradingProxyRancher(t *testing.T, provider string) error {
	return upgradeproxy.UpgradingProxyRancher(t, provider)
}

func UpgradeAirgapRancher(t *testing.T, client *rancher.Client, bastion, registry string, session *session.Session, cattleConfig map[string]any,
	tunnel *ssh.BastionSSHTunnel) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	return upgradeairgap.UpgradeAirgapRancher(t, client, bastion, registry, session, cattleConfig, tunnel)
}

func UpgradeDualStackRancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	return upgradedualstack.UpgradeDualStackRancher(t, client, serverNodeOne, session, cattleConfig)
}

func UpgradeIPv6Rancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	return upgradeipv6.UpgradeIPv6Rancher(t, client, serverNodeOne, session, cattleConfig)
}

func UpgradeProxyRancher(t *testing.T, client *rancher.Client, proxyPrivateIP, proxyBastion string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	return upgradeproxy.UpgradeProxyRancher(t, client, proxyPrivateIP, proxyBastion, session, cattleConfig)
}

func UpgradeRancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session,
	cattleConfig map[string]any) (*rancher.Client, map[string]any, *terraform.Options, *terraform.Options) {
	return upgradestandard.UpgradeRancher(t, client, serverNodeOne, session, cattleConfig)
}

func SetupResources(t *testing.T, client *rancher.Client, rancherConfig *rancher.Config, terratestConfig *config.TerratestConfig,
	terraformOptions *terraform.Options) (*rancher.Client, string, string, string) {
	return upgradebase.SetupResources(t, client, rancherConfig, terratestConfig, terraformOptions)
}

func CleanupDownstreamClusters(t *testing.T, client *rancher.Client, terraformConfig *config.TerraformConfig) {
	upgradebase.CleanupDownstreamClusters(t, client, terraformConfig)
}

func UniqueStrings(clusterIDs []string) []string {
	return upgradebase.UniqueStrings(clusterIDs)
}

func PostRancherSetup(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, session *session.Session, host,
	keyPath string, isUpgrade bool) (*rancher.Client, error) {
	return ranchersetup.PostRancherSetup(t, terraformOptions, rancherConfig, session, host, keyPath, isUpgrade)
}

func CreateAdminToken(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config) (*management.Token, error) {
	return ranchersetup.CreateAdminToken(t, terraformOptions, rancherConfig)
}

func CreateStandardUserToken(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, testUser, testPassword string) (*management.Token, error) {
	return ranchersetup.CreateStandardUserToken(t, terraformOptions, rancherConfig, testUser, testPassword)
}

func UpdateRancherConfigMap(cattleConfig map[string]any, client *rancher.Client) (map[string]any, error) {
	return ranchersetup.UpdateRancherConfigMap(cattleConfig, client)
}
