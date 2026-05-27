package setup

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tests/actions/features"
	featureDefaults "github.com/rancher/tfp-automation/framework/set/defaults/features"
)

// ToggleFeatureFlag applies the requested feature flag state.
func ToggleFeatureFlag(client *rancher.Client, feature string, toggledState string) {
	switch toggledState {
	case featureDefaults.ToggledOff, "true":
		features.UpdateFeatureFlag(client, feature, false)
	case featureDefaults.ToggledOn, "false":
		features.UpdateFeatureFlag(client, feature, true)
	}
}
