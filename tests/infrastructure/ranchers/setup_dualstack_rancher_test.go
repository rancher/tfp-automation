package ranchers

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework"
	"github.com/rancher/tfp-automation/framework/set/resources/dualstack"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DualStackRancherTestSuite struct {
	suite.Suite
	session          *session.Session
	terraformConfig  *config.TerraformConfig
	terratestConfig  *config.TerratestConfig
	standaloneConfig *config.Standalone
	cattleConfig     map[string]any
	rancherConfig    *rancher.Config
	terraformOptions *terraform.Options
}

func (i *DualStackRancherTestSuite) TestCreateRancher() {
	i.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	i.rancherConfig, i.terraformConfig, i.terratestConfig, i.standaloneConfig = config.LoadTFPConfigs(i.cattleConfig)

	_, keyPath := rancher2.SetKeyPath(keypath.DualStackKeyPath, i.terratestConfig.PathToRepo, i.terraformConfig.Provider)
	terraformOptions := framework.Setup(i.T(), i.terraformConfig, i.terratestConfig, keyPath)
	i.terraformOptions = terraformOptions

	_, err := dualstack.CreateMainTF(i.T(), i.terraformOptions, keyPath, i.rancherConfig, i.terraformConfig, i.terratestConfig)
	require.NoError(i.T(), err)

	testSession := session.NewSession()
	i.session = testSession

	_, err = PostRancherSetup(i.T(), i.terraformOptions, i.rancherConfig, i.session, i.terraformConfig.Standalone.RancherHostname, keyPath, false)
	require.NoError(i.T(), err)
}

func TestDualStackRancherTestSuite(t *testing.T) {
	suite.Run(t, new(DualStackRancherTestSuite))
}
