package permutationsdata

import (
	"strings"

	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/config/operations/permutations"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
)

const (
	k8sVersionKey = "kubernetesVersion"
)

// CreateK8sRelationships creates a relationship between the terraform module field and the terratest kubernetesVersion
func CreateK8sRelationships(cattleConfig map[string]any) ([]permutations.Relationship, error) {
	k8sKeyPath := []string{config.TerratestConfigurationFileKey, k8sVersionKey}
	k8sKeyValue, err := operations.GetValue(k8sKeyPath, cattleConfig)
	if err != nil {
		logrus.Warning("kubernetesVersion not set in config file")
		k8sKeyValue = []any{}
	}

	moduleKeyPath := []string{config.TerraformConfigurationFileKey, moduleKey}
	moduleKeyValue, err := operations.GetValue(moduleKeyPath, cattleConfig)
	if err != nil {
		return nil, err
	}

	var k8sRelationships []permutations.Relationship
	for _, module := range moduleKeyValue.([]any) {
		var k8sPermutation permutations.Permutation
		var k8sRelationship permutations.Relationship

		var k8sType string
		if strings.Contains(module.(string), "k3s") {
			k8sType = "k3s"
		} else if strings.Contains(module.(string), "rke2") {
			k8sType = "rke2"
		} else if strings.Contains(module.(string), "rke1") {
			k8sType = "rancher"
		}

		var matchedK8sVersions []any
		for _, k8sVersion := range k8sKeyValue.([]any) {
			if strings.Contains(k8sVersion.(string), k8sType) {
				matchedK8sVersions = append(matchedK8sVersions, k8sVersion)
			}
		}

		k8sPermutation = permutations.CreatePermutation(k8sKeyPath, matchedK8sVersions, nil)

		k8sRelationship = permutations.CreateRelationship(module, nil, nil, []permutations.Permutation{k8sPermutation})
		k8sRelationships = append(k8sRelationships, k8sRelationship)
	}

	return k8sRelationships, nil
}
