package resources

import (
	"os"
	"path/filepath"
)

// KeyPath is a function that will set the path to the key file.
func KeyPath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	keyPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/standalone/aws")

	return keyPath
}
