package main

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	infraConfig "github.com/rancher/tests/validation/recurring/infrastructure/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/tests/infrastructure"
	"github.com/sirupsen/logrus"
)

func main() {
	var client *rancher.Client
	var err error

	t := &testing.T{}

	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))

	client, _, _, _, _, err = setupRancher(t)
	if err != nil {
		logrus.Fatalf("Failed to setup Rancher: %v", err)
	}

	_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, client.RancherConfig.AdminToken, cattleConfig)
	if err != nil {
		logrus.Fatalf("Failed to replace admin token: %v", err)
	}

	infraConfig.WriteConfigToFile(os.Getenv(config.ConfigEnvironmentKey), cattleConfig)
}

func setupRancher(t *testing.T) (*rancher.Client, string, *terraform.Options, *terraform.Options, map[string]any, error) {
	testSession := session.NewSession()
	client, serverNodeOne, standaloneTerraformOptions, terraformOptions, cattleConfig := infrastructure.SetupRancher(t, testSession, keypath.SanityKeyPath)

	return client, serverNodeOne, standaloneTerraformOptions, terraformOptions, cattleConfig, nil
}
