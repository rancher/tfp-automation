package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/upgrade"
	"github.com/stretchr/testify/require"
)

// UpgradeAirgapRancher upgrades an existing airgapped Rancher server and returns the client, configuration, and Terraform options.
func UpgradeAirgapRancher(t *testing.T, client *rancher.Client, bastion, registry string, session *session.Session, cattleConfig map[string]any) (*rancher.Client,
	map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeAirgapRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, "", "", bastion, registry)
	require.NoError(t, err)

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}

// UpgradeProxyRancher upgrades an existing proxy Rancher server and returns the client, configuration, and Terraform options.
func UpgradeProxyRancher(t *testing.T, client *rancher.Client, proxyPrivateIP, proxyBastion string, session *session.Session, cattleConfig map[string]any) (*rancher.Client,
	map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeProxyRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, proxyPrivateIP, proxyBastion, "", "")
	require.NoError(t, err)

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}

// UpgradeRancher upgrades an existing Rancher server and returns the client, configuration, and Terraform options.
func UpgradeRancher(t *testing.T, client *rancher.Client, serverNodeOne string, session *session.Session, cattleConfig map[string]any) (*rancher.Client,
	map[string]any, *terraform.Options, *terraform.Options) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	terraformConfig.Standalone.UpgradeRancher = true

	_, keyPath := rancher2.SetKeyPath(keypath.UpgradeKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	upgradeTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	err = upgrade.CreateMainTF(t, upgradeTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig, serverNodeOne, "", "", "")
	require.NoError(t, err)

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	updatedCattleConfig, err := UpdateRancherConfigMap(cattleConfig, client)
	require.NoError(t, err)

	return client, updatedCattleConfig, terraformOptions, upgradeTerraformOptions
}
