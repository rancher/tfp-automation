package ranchers

import (
	"context"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/token"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// CreateAdminToken creates a new admin token for the Rancher client.
func CreateAdminToken(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config) (*management.Token, error) {
	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	var adminToken *management.Token
	err := kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.FiveMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		adminToken, err = token.GenerateUserToken(adminUser, rancherConfig.Host)
		if err != nil {
			logrus.Warnf("Failed to generate admin token: %v. Retrying...", err)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return adminToken, nil
}

// CreateStandardUserToken creates a new standard user token for the Rancher client.
func CreateStandardUserToken(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, testUser,
	testPassword string) (*management.Token, error) {
	standardUser := &management.User{
		Username: testUser,
		Password: testPassword,
	}

	var standardUserToken *management.Token
	err := kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.FiveMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		standardUserToken, err = token.GenerateUserToken(standardUser, rancherConfig.Host)
		if err != nil {
			logrus.Warnf("Failed to generate standard user token: %v. Retrying...", err)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return standardUserToken, nil
}
