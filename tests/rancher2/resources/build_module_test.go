package tests

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BuildModuleTestSuite struct {
	suite.Suite
	cattleConfig    map[string]any
	rancherConfig   *rancher.Config
	terraformConfig *config.TerraformConfig
	terratestConfig *config.TerratestConfig
}

func (r *BuildModuleTestSuite) TestBuildModule() {
	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	defer cleanup.TFFilesCleanup(keyPath)

	r.cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	r.rancherConfig, r.terraformConfig, r.terratestConfig = config.LoadTFPConfigs(r.cattleConfig)

	configMap := []map[string]any{r.cattleConfig}

	err := provisioning.BuildModule(r.T(), r.rancherConfig, r.terraformConfig, r.terratestConfig, configMap)
	require.NoError(r.T(), err)
}

func TestBuildModuleTestSuite(t *testing.T) {
	suite.Run(t, new(BuildModuleTestSuite))
}
