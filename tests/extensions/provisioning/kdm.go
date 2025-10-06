package provisioning

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/stretchr/testify/require"
)

// FetchSetting gets a Rancher global setting.
func FetchSetting(client *rancher.Client, settingName string) (string, error) {
	setting, err := client.Management.Setting.ByID(settingName)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

// Metadata is a struct for the KDM JSON payload extracting k8s versions.
type Metadata struct {
	Rke2 struct {
		Releases []struct {
			Version string `json:"version"`
		} `json:"releases"`
	} `json:"rke2"`
	K3s struct {
		Releases []struct {
			Version string `json:"version"`
		} `json:"releases"`
	} `json:"k3s"`
}

// VerifyKDMUrl validates the KDM URL and returns Kubernetes versions.
func VerifyKDMUrl(t *testing.T, url, rancherVersion string) map[string][]string {
	if strings.Contains(rancherVersion, "-alpha") || strings.Contains(rancherVersion, "-head") {
		require.True(t, strings.Contains(url, "dev-v"), "expected KDM URL to point to dev branch, url: %s", url)
	} else {
		require.True(t, strings.Contains(url, "release-v"), "expected KDM URL to point to release branch, url: %s", url)
	}

	resp, err := http.Get(url)
	require.NoError(t, err, "failed to GET KDM URL")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status code")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed reading KDM response body")
	require.NotEmpty(t, body, "KDM response body is empty")

	var meta Metadata
	err = json.Unmarshal(body, &meta)
	require.NoError(t, err, "failed to unmarshal KDM JSON")

	versions := make(map[string][]string)
	for _, rel := range meta.Rke2.Releases {
		if rel.Version != "" {
			versions["rke2"] = append(versions["rke2"], rel.Version)
		}
	}
	for _, rel := range meta.K3s.Releases {
		if rel.Version != "" {
			versions["k3s"] = append(versions["k3s"], rel.Version)
		}
	}

	return versions
}

// VerifyKDMVersions checks that the available Rancher Kubernetes versions are listed in the KDM version list
func VerifyKDMVersions(t *testing.T, kdmVersions map[string][]string, dropdownVersions []string, distro string) {
	var kdmList []string

	switch {
	case strings.Contains(distro, "rke2"):
		kdmList = kdmVersions["rke2"]
	case strings.Contains(distro, "k3s"):
		kdmList = kdmVersions["k3s"]
	default:
		t.Fatalf("Unsupported distro: %s", distro)
	}

	require.NotEmptyf(t, kdmList, "KDM version list for %s should not be empty", distro)
	require.NotEmptyf(t, dropdownVersions, "Dropdown version list for %s should not be empty", distro)

	for _, version := range dropdownVersions {
		require.Containsf(
			t,
			kdmList,
			version,
			"%s dropdown version %s not found in KDM versions: %v",
			distro,
			version,
			kdmList,
		)
	}
}
