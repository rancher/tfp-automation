package provisioning

import (
	"os"
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	framework "github.com/rancher/tfp-automation/framework/set"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
)

// BuildModule is a function that builds the Terraform module.
func BuildModule(t *testing.T, rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, configMap []map[string]any) error {
	keyPath := rancher2.SetKeyPath(keypath.RancherKeyPath)

	_, err := framework.ConfigTF(nil, "", "", "", configMap, false)
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
