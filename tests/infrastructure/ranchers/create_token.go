package ranchers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	v1Token "github.com/rancher/tfp-automation/tests/extensions/token"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// CreateAdminToken creates a new admin token for the Rancher client.
func CreateAdminToken(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config) (*management.Token, error) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	var adminToken *management.Token
	err := kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.FiveMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		if terraformConfig.GenerateV3Token {
			adminToken, err = token.GenerateUserToken(adminUser, rancherConfig.Host)
			if err != nil {
				logrus.Warnf("Failed to generate admin token: %v. Retrying...", err)
				return false, nil
			}
		} else {
			adminToken, err = v1Token.GenerateV1UserToken(adminUser, rancherConfig.Host)
			if err != nil {
				logrus.Warnf("Failed to generate admin token: %v. Retrying...", err)
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return adminToken, nil
}

// CreateStandardUserToken creates a new standard user token for the Rancher client.
func CreateStandardUserToken(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, testUser, testPassword string) (*management.Token, error) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	_, terraformConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	standardUser := &management.User{
		Username: testUser,
		Password: testPassword,
	}

	var standardUserToken *management.Token
	err := kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.FiveMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		if terraformConfig.GenerateV3Token {
			standardUserToken, err = token.GenerateUserToken(standardUser, rancherConfig.Host)
			if err != nil {
				logrus.Warnf("Failed to generate standard user token: %v. Retrying...", err)
				return false, nil
			}
		} else {
			standardUserToken, err = v1Token.GenerateV1UserToken(standardUser, rancherConfig.Host)
			if err != nil {
				logrus.Warnf("Failed to generate standard user token: %v. Retrying...", err)
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return standardUserToken, nil
}
