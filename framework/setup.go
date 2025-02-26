package framework

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/sirupsen/logrus"
)

// Setup is a function that will set the Terraform configuration and return the Terraform options.
func Setup(t *testing.T, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, keyPath string) *terraform.Options {
	var terratestLogger logger.Logger

	if strings.Contains(keyPath, keypath.RancherKeyPath) {
		terratestLogger = getLogger(terratestConfig.TFLogging)
	} else {
		terratestLogger = getLogger(terratestConfig.StandaloneLogging)
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: keyPath,
		NoColor:      true,
		Logger:       &terratestLogger,
	})

	return terraformOptions
}

func getLogger(tfLogging bool) logger.Logger {
	if tfLogging {
		logrus.Infof("Logging enabled. Terraform logs will be displayed.")
		return *logger.Default
	}

	logrus.Infof("Logging disabled. Terraform logs will be suppressed.")

	return *logger.Discard
}
