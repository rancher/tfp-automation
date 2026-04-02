package provisioning

import (
	"os"

	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
)

const (
	resourcePrefixKey = "resourcePrefix"
)

func UniquifyTerraform(cattleConfig map[string]any) (map[string]any, error) {
	resourcePrefix := []string{config.TerraformConfigurationFileKey, resourcePrefixKey}
	cattleConfig, err := uniquifyField(resourcePrefix, cattleConfig)
	if err != nil {
		return nil, err
	}

	return cattleConfig, nil
}

func uniquifyField(keyPath []string, cattleConfig map[string]any) (map[string]any, error) {
	cattleConfig = shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	keyPathValue := terraformConfig.ResourcePrefix

	keyPathValue = namegen.AppendRandomString(keyPathValue)

	uniqueCattleConfig, err := operations.ReplaceValue(keyPath, keyPathValue, cattleConfig)
	if err != nil {
		return nil, err
	}

	return uniqueCattleConfig, nil
}
