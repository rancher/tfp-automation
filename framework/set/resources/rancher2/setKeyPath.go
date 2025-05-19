package rancher2

import (
	"os"
	"path/filepath"
)

// SetKeyPath is a function that will set the path to the key file.
func SetKeyPath(keyPath, pathToRepo, provider string) (string, string) {
	var err error
	userDir := os.Getenv("GOPATH")
	if userDir == "" {
		userDir, err = os.UserHomeDir()
		if err != nil {
			return "", ""
		}

		userDir = filepath.Join(userDir, "go/")
	}

	keyPath = filepath.Join(userDir, pathToRepo, keyPath)

	if provider != "" {
		keyPath = filepath.Join(keyPath, "/", provider)
	}

	return userDir, keyPath
}
