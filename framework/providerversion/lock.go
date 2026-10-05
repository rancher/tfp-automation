package providerversion

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
)

const (
	mainTF   = "main.tf"
	lockFile = ".terraform.lock.hcl"
	pluginTF = ".terraform"
)

var (
	requiredProviderPattern = regexp.MustCompile(`(?s)rancher2\s*=\s*\{.*?version\s*=\s*"([^"]+)"`)
	lockedProviderPattern   = regexp.MustCompile(`(?s)provider\s+"registry\.terraform\.io/rancher/rancher2"\s*\{\s*version\s*=\s*"([^"]+)"`)
)

// ClearStaleLock removes the dependency lock when it pins a rancher2 provider the generated configuration no longer asks for.
func ClearStaleLock(terraformDir string) error {
	lockPath := filepath.Join(terraformDir, lockFile)

	locked, err := readProviderVersion(lockPath, lockedProviderPattern)
	if err != nil || locked == "" {
		return err
	}

	required, err := readProviderVersion(filepath.Join(terraformDir, mainTF), requiredProviderPattern)
	if err != nil || required == "" || required == locked {
		return err
	}

	logrus.Debugf("Removing the dependency lock in %s, it pins the rancher2 provider %s and the configuration asks for %s",
		terraformDir, locked, required)

	if err := os.Remove(lockPath); err != nil {
		return fmt.Errorf("removing the stale dependency lock %s: %w", lockPath, err)
	}

	if err := os.RemoveAll(filepath.Join(terraformDir, pluginTF)); err != nil {
		return fmt.Errorf("removing the plugin directory beside %s: %w", lockPath, err)
	}

	return nil
}

func readProviderVersion(path string, pattern *regexp.Regexp) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("reading %s to compare the rancher2 provider version: %w", path, err)
	}

	match := pattern.FindSubmatch(contents)
	if match == nil {
		return "", nil
	}

	return string(match[1]), nil
}
