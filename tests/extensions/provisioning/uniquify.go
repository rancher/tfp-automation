package provisioning

import (
	"github.com/rancher/shepherd/pkg/config/operations"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
)

const (
	resourcePrefixKey = "resourcePrefix"
)

func UniquifyTerraform(cattleConfigs []map[string]any) ([]map[string]any, error) {
	resourcePrefix := []string{config.TerraformConfigurationFileKey, resourcePrefixKey}
	var uniqueCattleConfigs []map[string]any
	for _, cattleConfig := range cattleConfigs {
		cattleConfig, err := uniquifyField(resourcePrefix, cattleConfig)
		if err != nil {
			return nil, err
		}

		uniqueCattleConfigs = append(uniqueCattleConfigs, cattleConfig)
	}

	return uniqueCattleConfigs, nil
}

func uniquifyField(keyPath []string, cattleConfig map[string]any) (map[string]any, error) {
	keyPathValue, err := operations.GetValue(keyPath, cattleConfig)
	if err != nil {
		return nil, err
	}

	keyPathValue = namegen.AppendRandomString(keyPathValue.(string))

	uniqueCattleConfig, err := operations.ReplaceValue(keyPath, keyPathValue, cattleConfig)
	if err != nil {
		return nil, err
	}

	return uniqueCattleConfig, nil
}
