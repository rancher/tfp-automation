package airgap

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/defaults/configs"
)

// ConfigAirgapCleanup is a function that will run terraform destroy and cleanup Terraform resources.
func ConfigAirgapCleanup(t *testing.T, terraformOptions *terraform.Options) {
	rancherConfig := new(rancher.Config)
	config.LoadConfig(configs.Rancher, rancherConfig)

	if *rancherConfig.Cleanup {
		terraform.Destroy(t, terraformOptions)
		ConfigAirgapCleanupTF()
	}
}
