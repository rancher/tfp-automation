package provisioning

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/sirupsen/logrus"
)

// BuildModule is a function that builds the Terraform module.
func BuildModule(t *testing.T) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	clusterConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(configs.Terratest, clusterConfig)

	keyPath := resources.SetKeyPath()

	err := framework.ConfigTF(nil, rancherConfig, terraformConfig, clusterConfig, "", "", "", "", "")
	if err != nil {
		return err
	}

	module, err := os.ReadFile(keyPath + configs.MainTF)
	if err != nil {
		logrus.Errorf("Failed to read main.tf file contents. Error: %v", err)

		return err
	}

	t.Log(string(module))

	return nil
}
