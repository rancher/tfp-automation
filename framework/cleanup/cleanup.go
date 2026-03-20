package cleanup

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/sirupsen/logrus"
)

// Cleanup is a function that will run terraform destroy and cleanup Terraform resources.
func Cleanup(t *testing.T, terraformOptions *terraform.Options, keyPath string) {
	rancherConfig := new(rancher.Config)
	config.LoadConfig(configs.Rancher, rancherConfig)

	if *rancherConfig.Cleanup {
		logrus.Infof("Cleaning up Terraform resources...")
		terraform.Destroy(t, terraformOptions)

		if !strings.Contains(keyPath, keypath.RancherKeyPath) {
			err := TFFilesCleanup(keyPath)
			if err != nil {
				logrus.Warning(err)
			}
		}
	}
}
