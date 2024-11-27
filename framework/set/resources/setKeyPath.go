package resources

import (
	"os"
	"path/filepath"
)

// SetKeyPath is a function that will set the path to the key file.
func SetKeyPath() string {
	var keyPath string
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	keyPath = filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/modules/rancher")

	return keyPath
}
