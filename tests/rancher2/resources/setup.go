package resources

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	configDefaults "github.com/rancher/tests/actions/config/defaults"
	"github.com/rancher/tests/validation/provisioning/resources/standarduser"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	ranchersetup "github.com/rancher/tfp-automation/tests/infrastructure/ranchers/setup"
	"github.com/stretchr/testify/require"
)

type resourcesSetup struct {
	Client             *rancher.Client
	StandardUserClient *rancher.Client
	Session            *session.Session
	CattleConfig       map[string]any
	RancherConfig      *rancher.Config
	TerraformConfig    *config.TerraformConfig
	TerratestConfig    *config.TerratestConfig
	TerraformOptions   *terraform.Options
	NodeRoles          []config.Nodepool
	StandardToken      string
	RKE2Module         string
	K3SModule          string
}

func Setup(t *testing.T) *resourcesSetup {
	r := &resourcesSetup{}

	var err error

	r.CattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

	r.CattleConfig, err = configDefaults.LoadPackageDefaults(r.CattleConfig, "")
	require.NoError(t, err)

	r.RancherConfig, r.TerraformConfig, r.TerratestConfig, _ = config.LoadTFPConfigs(r.CattleConfig)

	testSession := session.NewSession()
	r.Session = testSession

	_, keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath, r.TerratestConfig.PathToRepo, "")
	terraformOptions := framework.Setup(t, r.TerraformConfig, r.TerratestConfig, keyPath)

	r.TerraformOptions = terraformOptions

	client, err := ranchersetup.PostRancherSetup(t, r.TerraformOptions, r.RancherConfig, r.Session, r.RancherConfig.Host, keyPath, false)
	require.NoError(t, err)

	r.Client = client

	var testUser, testPassword string

	standardUserClient, testUser, testPassword, err := standarduser.CreateStandardUser(r.Client)
	require.NoError(t, err)

	r.StandardUserClient = standardUserClient

	standardUserToken, err := ranchersetup.CreateStandardUserToken(t, r.TerraformOptions, r.RancherConfig, testUser, testPassword)
	require.NoError(t, err)

	r.StandardToken = standardUserToken.Token

	r.NodeRoles = []config.Nodepool{config.EtcdNodePool, config.ControlPlaneNodePool, config.WorkerNodePool}
	r.RKE2Module, _, r.K3SModule = provisioning.DownstreamClusterModules(r.TerraformConfig)

	return r
}
