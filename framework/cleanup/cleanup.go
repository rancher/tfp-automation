package framework

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
)

// Cleanup is a function that will run terraform destroy and cleanup Terraform resources.
func Cleanup(t *testing.T, terraformOptions *terraform.Options) {
	rancherConfig := new(rancher.Config)
	config.LoadConfig("rancher", rancherConfig)

	if *rancherConfig.Cleanup {
		terraform.Destroy(t, terraformOptions)
		CleanupConfigTF()
	}
}
