package ranchers

import "github.com/rancher/shepherd/clients/rancher"

// UpdateRancherConfigMap updates the rancher map in the cattleConfig with the latest Rancher client config values.
func UpdateRancherConfigMap(cattleConfig map[string]any, client *rancher.Client) (map[string]any, error) {
	if rancherMap, ok := cattleConfig["rancher"].(map[string]any); ok {
		rancherMap["host"] = client.RancherConfig.Host
		rancherMap["adminToken"] = client.RancherConfig.AdminToken
		rancherMap["adminPassword"] = client.RancherConfig.AdminPassword
		cattleConfig["rancher"] = rancherMap
	}

	return cattleConfig, nil
}
