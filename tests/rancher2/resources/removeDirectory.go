package resources

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/sirupsen/logrus"
)

// RemoveDirectory is a helper function that will remove the specified directory and report an error if it fails.
func RemoveDirectory(t *testing.T, rancherConfig *rancher.Config, nestedRancherModuleDir string) {
	if t.Failed() || rancherConfig.Cleanup == nil || !*rancherConfig.Cleanup {
		logrus.Infof("Preserving Terraform directory: %s", nestedRancherModuleDir)
		return
	}

	err := os.RemoveAll(nestedRancherModuleDir)
	if err != nil {
		t.Errorf("Failed to remove directory %s: %v", nestedRancherModuleDir, err)
	}
}
