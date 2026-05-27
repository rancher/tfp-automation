package ipv6

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	resources "github.com/rancher/tfp-automation/framework/set/resources/ipv6"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rancherinternal "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/internal"
	ranchersetup "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// UpgradingIPv6Rancher is a function that creates and upgrades an IPv6 Rancher setup, either via CLI or web application
func UpgradingIPv6Rancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.IPv6KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	serverNodeOne, err := resources.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	testSession := session.NewSession()

	terraformConfig.Standalone.UpgradeIPv6Rancher = true

	_, upgradeKeyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, upgradeKeyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, upgradeKeyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)

	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keypath.IPv6KeyPath)

	_, err = ranchersetup.PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	return nil
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
	client, err = ranchersetup.PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, true)
	require.NoError(t, err)

	updatedCattleConfig, err := ranchersetup.UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	if standaloneConfig.UpgradedRancherTagVersion != rancherinternal.Head {
		_, ipv6KeyPath := rancher2.SetKeyPath(keypath.IPv6KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
		terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, ipv6KeyPath)

		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.UpgradedRancherTagVersion, ipv6KeyPath, terraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetLocal + "/local")
	require.NoError(t, err)

	logrus.Infof("Verifying cluster deployments (%s)", cluster.Name)
	err = deployment.VerifyClusterDeployments(client, cluster)
	require.NoError(t, err)

	logrus.Infof("Verifying cluster pods (%s)", cluster.Name)
	err = pods.VerifyClusterPods(client, cluster)
	require.NoError(t, err)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}
