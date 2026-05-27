package standard

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	configDefaults "github.com/rancher/tests/actions/config/defaults"
	"github.com/rancher/tests/actions/features"
	"github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tests/actions/workloads/pods"
	infraConfig "github.com/rancher/tests/validation/recurring/infrastructure/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	featureDefaults "github.com/rancher/tfp-automation/framework/set/defaults/features"
	"github.com/rancher/tfp-automation/framework/set/resources/hosted"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rancherinternal "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/internal"
	ranchersetup "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateRancher is a function that creates a Rancher setup, either via CLI or web application
func CreateRancher(t *testing.T, provider string) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	configPath := os.Getenv("CATTLE_TEST_CONFIG")
	cattleConfig := shepherdConfig.LoadConfigFromFile(configPath)
	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.SanityKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, err := sanity.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	// For providers that do not have built-in DNS records, this will update the Rancher server URL.
	if terraformConfig.Provider != providers.AWS && terraformConfig.Provider != providers.Azure {
		_, err = operations.ReplaceValue([]string{"rancher", "host"}, terraformConfig.Standalone.RancherHostname, cattleConfig)
		require.NoError(t, err)

		rancherConfig, terraformConfig, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)
		infraConfig.WriteConfigToFile(os.Getenv(rancherinternal.ConfigEnvironmentKey), cattleConfig)
	}

	var client *rancher.Client
	testSession := session.NewSession()

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.MCM == "" {
		client, err = ranchersetup.PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
		if err != nil {
			return err
		}
	} else if standaloneConfig.FeatureFlags == nil {
		client, err = ranchersetup.PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
		if err != nil {
			return err
		}
	}

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
		ranchersetup.ToggleFeatureFlag(client, featureDefaults.Turtles, standaloneConfig.FeatureFlags.Turtles)
	}

	return nil
}

// SetupRancher sets up a Rancher server and returns the client, configuration, and Terraform options.
func SetupRancher(t *testing.T, session *session.Session, moduleKeyPath string) (*rancher.Client, string, *terraform.Options,
	*terraform.Options, map[string]any) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

	var err error
	cattleConfig, err = configDefaults.LoadPackageDefaults(cattleConfig, rancherinternal.DefaultsFilePath)
	require.NoError(t, err)

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	var serverNodeOne string

	if terraformConfig.LocalHostedCluster {
		serverNodeOne, err = hosted.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
		require.NoError(t, err)
	} else {
		serverNodeOne, err = sanity.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
		require.NoError(t, err)
	}

	// For providers that do not have built-in DNS records, this will update the Rancher server URL.
	if terraformConfig.Provider != providers.AWS && terraformConfig.Provider != providers.Azure && terraformConfig.Provider != providers.EKS &&
		terraformConfig.Provider != providers.AKS {
		_, err = operations.ReplaceValue([]string{"rancher", "host"}, terraformConfig.Standalone.RancherHostname, cattleConfig)
		require.NoError(t, err)

		_, err = operations.ReplaceValue([]string{"terraform", "standalone", "rancherHostname"}, terraformConfig.Standalone.RancherHostname, cattleConfig)
		require.NoError(t, err)

		rancherConfig, terraformConfig, terratestConfig, _ = config.LoadTFPConfigs(cattleConfig)
		infraConfig.WriteConfigToFile(os.Getenv(rancherinternal.ConfigEnvironmentKey), cattleConfig)
	}

	client, err := ranchersetup.PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != rancherinternal.Head {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	if standaloneConfig.FeatureFlags != nil && standaloneConfig.FeatureFlags.Turtles != "" {
		switch standaloneConfig.FeatureFlags.Turtles {
		case featureDefaults.ToggledOff:
			features.UpdateFeatureFlag(client, featureDefaults.Turtles, false)
		case featureDefaults.ToggledOn:
			features.UpdateFeatureFlag(client, featureDefaults.Turtles, true)
		}
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

	return client, serverNodeOne, standaloneTerraformOptions, terraformOptions, cattleConfig
}
