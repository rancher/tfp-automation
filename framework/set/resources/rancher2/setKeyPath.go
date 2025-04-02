package rancher2

import (
	"os"
	"path/filepath"

	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
)

// SetKeyPath is a function that will set the path to the key file.
func SetKeyPath(keyPath string, terraformConfig *config.TerraformConfig) string {
	var err error
	userDir := os.Getenv("GOROOT")
	if userDir == "" {
		userDir, err = os.UserHomeDir()
		if err != nil {
			return ""
		}
	}

	if terraformConfig == nil {
		return filepath.Join(userDir, keyPath)
	}

	switch terraformConfig.NodeProvider {
	case defaults.Aws:
		keyPath = filepath.Join(userDir, keyPath, "/aws")
	case defaults.Linode:
		keyPath = filepath.Join(userDir, keyPath, "/linode")
	default:
		logrus.Errorf("Unsupported provider: %s", terraformConfig.NodeProvider)
	}

	return keyPath
}
