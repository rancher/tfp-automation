package tests

import (
	"testing"

	cleanup "github.com/rancher/tfp-automation/framework/cleanup/rancher2"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BuildModuleTestSuite struct {
	suite.Suite
}

func (r *BuildModuleTestSuite) TestBuildModule() {
	defer cleanup.ConfigCleanupTF()

	err := provisioning.BuildModule(r.T())
	require.NoError(r.T(), err)
}

func TestBuildModuleTestSuite(t *testing.T) {
	suite.Run(t, new(BuildModuleTestSuite))
}
