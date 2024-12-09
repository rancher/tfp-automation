package rke

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	mainTfKeyPath = "RKE_KEY_PATH"
)

// KeyPath is a function that will set the path to the key file.
func KeyPath() string {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	mainTfDirPath := os.Getenv(mainTfKeyPath)
	if mainTfDirPath == "" {
		logrus.Fatalf("Expected env var not set: %s", mainTfKeyPath)
	}

	keyPath := filepath.Join(userDir, mainTfDirPath)

	return keyPath
}
