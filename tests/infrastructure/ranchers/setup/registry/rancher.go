package registry

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/namespaces"
	"github.com/rancher/shepherd/pkg/session"
	reg "github.com/rancher/tests/actions/registries"
	"github.com/rancher/tests/actions/workloads/deployment"
	"github.com/rancher/tests/actions/workloads/pods"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/stevetypes"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/registries"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	rancherinternal "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/internal"
	ranchersetup "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// CreateRegistryRancher is a function that creates a registry-enabled Rancher setup, either via CLI or web application
func CreateRegistryRancher(t *testing.T, provider string, cattleConfig map[string]any) error {
	os.Getenv("CLOUD_PROVIDER_VERSION")

	rancherConfig, terraformConfig, terratestConfig, _ := config.LoadTFPConfigs(cattleConfig)

	if provider != "" {
		terraformConfig.Provider = provider
	}

	_, keyPath := rancher2.SetKeyPath(keypath.RegistryKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	_, _, _, err := registries.CreateMainTF(t, terraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	if err != nil {
		return err
	}

	testSession := session.NewSession()

	_, err = ranchersetup.PostRancherSetup(t, terraformOptions, rancherConfig, testSession, terraformConfig.Standalone.RancherHostname, keyPath, false)
	if err != nil {
		return err
	}

	return nil
}

// SetupRegistryRancher sets up a registry-enabled Rancher server and returns the client, configuration, and Terraform options.
func SetupRegistryRancher(t *testing.T, session *session.Session, moduleKeyPath string, cattleConfig map[string]any) (*rancher.Client, string, string, string,
	*terraform.Options, *terraform.Options, map[string]any) {
	var err error

	rancherConfig, terraformConfig, terratestConfig, standaloneConfig := config.LoadTFPConfigs(cattleConfig)

	_, keyPath := rancher2.SetKeyPath(moduleKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)
	standaloneTerraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	authRegistry, unauthRegistry, globalRegistry, err := registries.CreateMainTF(t, standaloneTerraformOptions, keyPath, rancherConfig, terraformConfig, terratestConfig)
	require.NoError(t, err)

	client, err := ranchersetup.PostRancherSetup(t, standaloneTerraformOptions, rancherConfig, session, terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(t, err)

	if standaloneConfig.RancherTagVersion != rancherinternal.Head {
		provisioning.VerifyRancherVersion(t, rancherConfig.Host, standaloneConfig.RancherTagVersion, keyPath, standaloneTerraformOptions)
	}

	_, keyPath = rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, terraformConfig, terratestConfig, keyPath)

	usesRegistryPrefix, err := reg.CheckAllClusterPodsForRegistryPrefix(client, rancherinternal.Local, globalRegistry)
	require.NoError(t, err)

	if !usesRegistryPrefix {
		t.Fatalf("ERROR: not all of the local cluster pods are using the global registry")
	}

	cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(namespaces.FleetLocal + "/local")
	require.NoError(t, err)

	logrus.Infof("Verifying cluster deployments (%s)", cluster.Name)
	err = deployment.VerifyClusterDeployments(client, cluster)
	require.NoError(t, err)

	logrus.Infof("Verifying cluster pods (%s)", cluster.Name)
	err = pods.VerifyClusterPods(client, cluster)
	require.NoError(t, err)

	return client, authRegistry, unauthRegistry, globalRegistry, standaloneTerraformOptions, terraformOptions, cattleConfig
}
