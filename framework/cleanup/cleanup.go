package cleanup

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/defaults/configs"
)

// Cleanup is a function that will run terraform destroy and cleanup Terraform resources.
func Cleanup(t *testing.T, terraformOptions *terraform.Options, keyPath string) {
	rancherConfig := new(rancher.Config)
	config.LoadConfig(configs.Rancher, rancherConfig)

	if *rancherConfig.Cleanup {
		terraform.Destroy(t, terraformOptions)
		TFFilesCleanup(keyPath)
	}
}
