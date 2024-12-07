package tests

import (
	"testing"

	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/stretchr/testify/suite"
)

type CleanupTestSuite struct {
	suite.Suite
}

func (r *CleanupTestSuite) TestCleanup() {
	provisioning.ForceCleanup(r.T())
}

func TestCleanupTestSuite(t *testing.T) {
	suite.Run(t, new(CleanupTestSuite))
}
