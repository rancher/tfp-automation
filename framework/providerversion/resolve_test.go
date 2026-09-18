package providerversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderMajor(t *testing.T) {
	cases := map[string]int{"2.6": 2, "2.8": 4, "2.12": 8, "2.13": 13, "2.15": 15, "2.16": 16}

	for minor, expected := range cases {
		major, ok := providerMajor(minor)
		require.True(t, ok, "expected a provider major for Rancher %s", minor)
		assert.Equal(t, expected, major, "wrong provider major for Rancher %s", minor)
	}

	_, ok := providerMajor("2.5")
	assert.False(t, ok, "Rancher 2.5 predates the mapping")

	_, ok = providerMajor("3.0")
	assert.False(t, ok, "only Rancher 2.x is mapped")
}

func TestNewestForMajor(t *testing.T) {
	candidates := []string{"8.0.0", "8.5.0", "8.3.1", "13.1.4", "16.0.0-mainhead", "16.0.0-rc.1"}

	assert.Equal(t, "8.5.0", newestForMajor(candidates, 8))
	assert.Equal(t, "16.0.0-mainhead", newestForMajor(candidates, 16), "rc builds are not served by a registry mirror")
	assert.Equal(t, "", newestForMajor(candidates, 14))
	assert.Equal(t, "13.1.4", newestBelowMajor(candidates, 14))
}

func TestNewestPrefersStable(t *testing.T) {
	assert.Equal(t, "15.0.0", newest([]string{"15.0.0-headrevert", "15.0.0", "15.0.0-v15head"}))
	assert.Equal(t, "15.1.2", newest([]string{"15.0.0", "15.1.2"}))
}

func TestMirrorDiscovery(t *testing.T) {
	mirror := mirrorDirectory()
	if mirror == "" {
		t.Skip("no filesystem mirror configured on this machine")
	}

	versions, err := mirroredVersions(mirror)
	require.NoError(t, err)

	t.Logf("mirror %s serves %d versions for this platform: %v", mirror, len(versions), versions)
	assert.NotEmpty(t, versions)
}
