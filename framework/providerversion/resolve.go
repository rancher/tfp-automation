package providerversion

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	ProviderEnvVar = "RANCHER2_PROVIDER_VERSION"

	providerAddress     = "registry.terraform.io/rancher/rancher2"
	registryVersionURL  = "https://registry.terraform.io/v1/providers/rancher/rancher2/versions"
	serverVersionPath   = "/v3/settings/server-version"
	terraformRC         = ".terraformrc"
	cliConfigEnvVar     = "TF_CLI_CONFIG_FILE"
	defaultMirror       = ".terraform.d/plugins"
	localSourceMarker   = "-rc"
	alignedRancherMinor = 13
	requestTimeout      = 30 * time.Second
)

var legacyProviderMajors = map[int]int{6: 2, 7: 3, 8: 4, 9: 5, 10: 6, 11: 7, 12: 8}

var rancherMinorPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)`)

var versionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)

var (
	resolvedLock sync.Mutex
	resolved     = map[string]string{}
)

// Resolve returns the rancher2 provider version to pin for the Rancher under test.
func Resolve(host, adminToken string, insecure bool, pinned string, pinnedByMinor map[string]string) (string, error) {
	resolvedLock.Lock()
	defer resolvedLock.Unlock()

	if version, ok := resolved[host]; ok {
		return version, nil
	}

	if version := strings.TrimPrefix(os.Getenv(ProviderEnvVar), "v"); version != "" {
		resolved[host] = version
		logrus.Infof("Using rancher2 provider version %s, from %s", version, ProviderEnvVar)

		return version, nil
	}

	if pinned != "" {
		version := strings.TrimPrefix(pinned, "v")
		resolved[host] = version
		logrus.Infof("Using rancher2 provider version %s, from the config file", version)

		return version, nil
	}

	version, err := resolveForServer(host, adminToken, insecure, pinnedByMinor)
	if err != nil {
		return "", err
	}

	resolved[host] = version
	logrus.Infof("Using rancher2 provider version %s, resolved for the Rancher under test", version)

	return version, nil
}

// resolveForServer picks the newest installable provider aligned with the Rancher under test.
func resolveForServer(host, adminToken string, insecure bool, pinnedByMinor map[string]string) (string, error) {
	serverVersion, err := rancherServerVersion(host, adminToken, insecure)
	if err != nil {
		return "", fmt.Errorf("could not read the Rancher server version to pick a provider, so set %s: %w", ProviderEnvVar, err)
	}

	minor := rancherMinorPattern.FindStringSubmatch(serverVersion)
	if minor == nil {
		return "", fmt.Errorf("could not read a minor version out of Rancher %s, so set %s", serverVersion, ProviderEnvVar)
	}

	minorKey := minor[1] + "." + minor[2]

	if version, ok := pinnedByMinor[minorKey]; ok {
		logrus.Infof("Using the rancher2 provider %s pinned for Rancher %s", version, minorKey)

		return strings.TrimPrefix(version, "v"), nil
	}

	major, ok := providerMajor(minorKey)
	if !ok {
		return "", fmt.Errorf("no rancher2 provider major is known for Rancher %s, so set %s or pin it under terraform.providerVersions", minorKey, ProviderEnvVar)
	}

	candidates, origin, err := installableVersions()
	if err != nil {
		return "", fmt.Errorf("could not list the installable rancher2 provider versions, so set %s: %w", ProviderEnvVar, err)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no rancher2 provider versions are installable from %s, so set %s", origin, ProviderEnvVar)
	}

	version := newestForMajor(candidates, major)
	if version != "" {
		logrus.Infof("Resolved the rancher2 provider %s for Rancher %s from %s", version, serverVersion, origin)

		return version, nil
	}

	version = newestBelowMajor(candidates, major)
	if version == "" {
		return "", fmt.Errorf("none of the rancher2 provider versions in %s are usable with Rancher %s, so set %s", origin, serverVersion, ProviderEnvVar)
	}

	logrus.Warnf("No rancher2 provider %d.x is installable from %s for Rancher %s, falling back to %s; set %s to override", major, origin, serverVersion, version, ProviderEnvVar)

	return version, nil
}

// rancherServerVersion reads the server-version setting from the Rancher under test.
func rancherServerVersion(host, adminToken string, insecure bool) (string, error) {
	request, err := http.NewRequest(http.MethodGet, "https://"+strings.TrimSuffix(host, "/")+serverVersionPath, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: requestTimeout}
	if insecure {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s returned %s", serverVersionPath, response.Status)
	}

	setting := struct {
		Value string `json:"value"`
	}{}

	err = json.NewDecoder(response.Body).Decode(&setting)
	if err != nil {
		return "", err
	}

	if setting.Value == "" {
		return "", fmt.Errorf("%s returned an empty value", serverVersionPath)
	}

	return setting.Value, nil
}

// providerMajor maps a Rancher minor version onto the rancher2 provider major that tracks it.
func providerMajor(rancherMinor string) (int, bool) {
	parts := strings.Split(rancherMinor, ".")
	if len(parts) != 2 || parts[0] != "2" {
		return 0, false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}

	major, ok := legacyProviderMajors[minor]
	if ok {
		return major, true
	}

	if minor >= alignedRancherMinor {
		return minor, true
	}

	return 0, false
}

// installableVersions lists the provider versions this machine can actually install.
func installableVersions() ([]string, string, error) {
	mirror := mirrorDirectory()
	if mirror != "" {
		versions, err := mirroredVersions(mirror)
		if err != nil {
			return nil, mirror, err
		}

		return versions, mirror, nil
	}

	versions, err := registryVersions()

	return versions, "the terraform registry", err
}

// cliConfigPath returns the terraform CLI configuration file, if one exists.
func cliConfigPath() string {
	path := os.Getenv(cliConfigEnvVar)
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}

		path = filepath.Join(home, terraformRC)
	}

	_, err := os.Stat(path)
	if err != nil {
		return ""
	}

	return path
}

// mirrorDirectory returns the filesystem mirror serving the rancher2 provider, if one is configured.
func mirrorDirectory() string {
	path := cliConfigPath()
	if path == "" {
		return defaultMirrorDirectory()
	}

	parsed, diagnostics := hclparse.NewParser().ParseHCLFile(path)
	if diagnostics.HasErrors() || parsed == nil {
		return defaultMirrorDirectory()
	}

	content, _, _ := parsed.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "provider_installation"}},
	})

	for _, installation := range content.Blocks {
		mirrors, _, _ := installation.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "filesystem_mirror"}},
		})

		for _, mirror := range mirrors.Blocks {
			attributes, _ := mirror.Body.JustAttributes()

			if !mirrorServesProvider(attributes) {
				continue
			}

			pathAttribute, ok := attributes["path"]
			if !ok {
				continue
			}

			value, diagnostics := pathAttribute.Expr.Value(nil)
			if diagnostics.HasErrors() || value.Type() != cty.String {
				continue
			}

			return value.AsString()
		}
	}

	return defaultMirrorDirectory()
}

// mirrorServesProvider reports whether a filesystem_mirror block covers the rancher2 provider.
func mirrorServesProvider(attributes hcl.Attributes) bool {
	include, ok := attributes["include"]
	if !ok {
		return true
	}

	value, diagnostics := include.Expr.Value(nil)
	if diagnostics.HasErrors() || !value.CanIterateElements() {
		return true
	}

	for iterator := value.ElementIterator(); iterator.Next(); {
		_, element := iterator.Element()
		if element.Type() != cty.String {
			continue
		}

		if matchesProvider(element.AsString()) {
			return true
		}
	}

	return false
}

// matchesProvider reports whether a mirror include pattern covers the rancher2 provider address.
func matchesProvider(pattern string) bool {
	if pattern == providerAddress {
		return true
	}

	prefix, found := strings.CutSuffix(pattern, "*")

	return found && strings.HasPrefix(providerAddress, prefix)
}

// defaultMirrorDirectory returns the conventional plugin directory when it holds the rancher2 provider.
func defaultMirrorDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(home, defaultMirror)

	_, err = os.Stat(filepath.Join(path, providerAddress))
	if err != nil {
		return ""
	}

	return path
}

// mirroredVersions lists the mirrored provider versions that carry a build for this platform.
func mirroredVersions(mirror string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(mirror, providerAddress))
	if err != nil {
		return nil, err
	}

	platform := runtime.GOOS + "_" + runtime.GOARCH

	var versions []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		_, err = os.Stat(filepath.Join(mirror, providerAddress, entry.Name(), platform))
		if err != nil {
			continue
		}

		versions = append(versions, entry.Name())
	}

	return versions, nil
}

// registryVersions lists the provider versions published to the terraform registry.
func registryVersions() ([]string, error) {
	client := &http.Client{Timeout: requestTimeout}

	response, err := client.Get(registryVersionURL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the terraform registry returned %s", response.Status)
	}

	payload := struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}{}

	err = json.NewDecoder(response.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	var versions []string

	for _, entry := range payload.Versions {
		versions = append(versions, entry.Version)
	}

	return versions, nil
}

// newestForMajor returns the newest usable version carrying the given major.
func newestForMajor(candidates []string, major int) string {
	var matching []string

	for _, candidate := range candidates {
		parts := versionPattern.FindStringSubmatch(candidate)
		if parts == nil || strings.Contains(candidate, localSourceMarker) {
			continue
		}

		candidateMajor, _ := strconv.Atoi(parts[1])
		if candidateMajor == major {
			matching = append(matching, candidate)
		}
	}

	return newest(matching)
}

// newestBelowMajor returns the newest usable version carrying a major below the given one.
func newestBelowMajor(candidates []string, major int) string {
	var matching []string

	for _, candidate := range candidates {
		parts := versionPattern.FindStringSubmatch(candidate)
		if parts == nil || strings.Contains(candidate, localSourceMarker) {
			continue
		}

		candidateMajor, _ := strconv.Atoi(parts[1])
		if candidateMajor < major {
			matching = append(matching, candidate)
		}
	}

	return newest(matching)
}

// newest returns the highest version, preferring a stable release over a prerelease.
func newest(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	sort.Slice(versions, func(i, j int) bool {
		return less(versions[i], versions[j])
	})

	return versions[len(versions)-1]
}

// less orders two provider versions numerically, ranking a prerelease below its stable release.
func less(left, right string) bool {
	leftParts := versionPattern.FindStringSubmatch(left)
	rightParts := versionPattern.FindStringSubmatch(right)

	for index := 1; index <= 3; index++ {
		leftNumber, _ := strconv.Atoi(leftParts[index])
		rightNumber, _ := strconv.Atoi(rightParts[index])

		if leftNumber != rightNumber {
			return leftNumber < rightNumber
		}
	}

	if (leftParts[4] == "") != (rightParts[4] == "") {
		return leftParts[4] != ""
	}

	return leftParts[4] < rightParts[4]
}
