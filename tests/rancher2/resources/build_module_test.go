package tests

import (
	"testing"

	"github.com/rancher/tfp-automation/defaults/keypath"
	cleanup "github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BuildModuleTestSuite struct {
	suite.Suite
}

func (r *BuildModuleTestSuite) TestBuildModule() {
	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)
	defer cleanup.TFFilesCleanup(keyPath)

	err := provisioning.BuildModule(r.T())
	require.NoError(r.T(), err)
}

func TestBuildModuleTestSuite(t *testing.T) {
	suite.Run(t, new(BuildModuleTestSuite))
}
