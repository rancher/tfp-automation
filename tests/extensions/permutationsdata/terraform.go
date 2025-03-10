package permutationsdata

import (
	"strings"

	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/config/operations/permutations"
	"github.com/rancher/tfp-automation/config"
)

const (
	moduleKey    = "module"
	cniKey       = "cni"
	awsConfigKey = "awsConfig"
	amiKey       = "ami"
)

func CreateModulePermutation(cattleConfig map[string]any) (*permutations.Permutation, error) {
	moduleKeyPath := []string{config.TerraformConfigurationFileKey, moduleKey}
	moduleKeyValue, err := operations.GetValue(moduleKeyPath, cattleConfig)
	if err != nil {
		return nil, err
	}

	if _, ok := moduleKeyValue.([]any); !ok {
		moduleKeyValue = []any{moduleKeyValue}
	}
	modulePermutation := permutations.CreatePermutation(moduleKeyPath, moduleKeyValue.([]any), nil)

	return &modulePermutation, nil
}

func CreateCNIPermutation(cattleConfig map[string]any) (*permutations.Permutation, error) {
	cniKeyPath := []string{config.TerraformConfigurationFileKey, cniKey}
	cniKeyValue, err := operations.GetValue(cniKeyPath, cattleConfig)
	if err != nil {
		return nil, err
	}

	if _, ok := cniKeyValue.([]any); !ok {
		cniKeyValue = []any{cniKeyValue}
	}
	cniPermutation := permutations.CreatePermutation(cniKeyPath, cniKeyValue.([]any), nil)

	return &cniPermutation, nil
}

func createAMIPermutation(cattleConfig map[string]any) (*permutations.Permutation, error) {
	amiKeyPath := []string{config.TerraformConfigurationFileKey, awsConfigKey, amiKey}
	amiKeyValue, err := operations.GetValue(amiKeyPath, cattleConfig)
	amiPermutation := permutations.CreatePermutation(amiKeyPath, amiKeyValue.([]any), nil)

	return &amiPermutation, err
}

func CreateAMIRelationships(cattleConfig map[string]any) ([]permutations.Relationship, error) {
	moduleKeyPath := []string{config.TerraformConfigurationFileKey, moduleKey}
	moduleKeyValue, err := operations.GetValue(moduleKeyPath, cattleConfig)
	if err != nil {
		return nil, err
	}

	if _, ok := moduleKeyValue.([]any); !ok {
		moduleKeyValue = []any{moduleKeyValue}
	}

	amiPermutation, err := createAMIPermutation(cattleConfig)
	if err != nil {
		return nil, err
	}

	var amiRelationships []permutations.Relationship
	for _, module := range moduleKeyValue.([]any) {
		if !strings.Contains(module.(string), "ec2") {
			continue
		}

		amiRelationship := permutations.CreateRelationship(module, nil, nil, []permutations.Permutation{*amiPermutation})
		amiRelationships = append(amiRelationships, amiRelationship)
	}

	return amiRelationships, err
}
