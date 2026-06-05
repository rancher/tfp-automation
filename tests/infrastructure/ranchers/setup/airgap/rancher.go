package airgap

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/features"
	reg "github.com/rancher/tests/actions/registries"
	"github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tests/actions/workloads/pods"
	infraConfig "github.com/rancher/tests/validation/recurring/infrastructure/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	featureDefaults "github.com/rancher/tfp-automation/framework/set/defaults/features"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/tests/extensions/ssh"
	rancherinternal "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/internal"
	ranchersetup "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateAirgapRancher is a function that creates an airgap Rancher setup, either via CLI or web application
func CreateAirgapRancher(t *testing.T, provider string, cattleConfig map[string]any) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.AirgapKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, bastion, err := airgap.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	_, err = operations.ReplaceValue([]string{"terraform", "airgapBastion"}, terraformConfig.AirgapBastion, cattleConfig)
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "systemDefaultRegistry"}, terraformConfig.PrivateRegistries.SystemDefaultRegistry, cattleConfig)
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"terraform", "privateRegistries", "url"}, terraformConfig.PrivateRegistries.URL, cattleConfig)
	require.NoError(t, err)

	rancherConfig, terraformConfig, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)
	infraConfig.WriteConfigToFile(os.Getenv(rancherinternal.ConfigEnvironmentKey), cattleConfig)

	sshKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	require.NoError(t, err)

	_, err = ssh.StartBastionSSHTunnel(bastion, terraformConfig.Standalone.OSUser, sshKey, "8443", standaloneConfig.RancherHostname, "443")
	require.NoError(t, err)

	testSession := session.NewSession()

	client, err := ranchersetup.PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, client.RancherConfig.AdminToken, cattleConfig)
	require.NoError(t, err)

	rancherConfig, terraformConfig, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)
	infraConfig.WriteConfigToFile(os.Getenv(rancherinternal.ConfigEnvironmentKey), cattleConfig)

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
		switch standaloneConfig.FeatureFlags.Turtles {
		case featureDefaults.ToggledOff:
			features.UpdateFeatureFlag(client, featureDefaults.Turtles, false)
		case featureDefaults.ToggledOn:
			features.UpdateFeatureFlag(client, featureDefaults.Turtles, true)
		}
	}

	return nil
}

// SetupAirgapRancher sets up an airgapped Rancher server and returns the client, configuration, and Terraform options.
func SetupAirgapRancher(t *testing.T, session *session.Session, moduleKeyPath string, cattleConfig map[string]any) (*rancher.Client, string, string, *terraform.Options,
	*terraform.Options, map[string]any, *ssh.BastionSSHTunnel) {
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	registry, bastion, err := airgap.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	sshKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	require.NoError(t, err)

	tunnel, err := ssh.StartBastionSSHTunnel(bastion, terraformConfig.Standalone.OSUser, sshKey, "8443", standaloneConfig.RancherHostname, "443")
	require.NoError(t, err)

	client, err := ranchersetup.PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(t, err)

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	usesRegistryPrefix, err := reg.CheckAllClusterPodsForRegistryPrefix(client, rancherinternal.Local, registry)
	require.NoError(t, err)

	if !usesRegistryPrefix {
		t.Fatalf("ERROR: not all of the local cluster pods are using the private registry")
	}

	cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetLocal + "/local")
	require.NoError(t, err)

	logrus.Infof("Verifying cluster deployments (%s)", cluster.Name)
	err = deployment.VerifyClusterDeployments(client, cluster)
	require.NoError(t, err)

	logrus.Infof("Verifying cluster pods (%s)", cluster.Name)
	err = pods.VerifyClusterPods(client, cluster)
	require.NoError(t, err)

	return client, registry, bastion, standaloneTerraformOptions, terraformOptions, cattleConfig, tunnel
}
