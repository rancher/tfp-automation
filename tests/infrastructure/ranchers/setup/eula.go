package setup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/pipeline"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// PostRancherSetup is a helper function that creates a Rancher client and accepts the EULA, if needed
func PostRancherSetup(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, session *session.Session, host,
	keyPath string, isUpgrade bool) (*rancher.Client, error) {
	adminToken, err := CreateAdminToken(t, terraformOptions, rancherConfig)
	if err != nil {
		if *rancherConfig.Cleanup {
			logrus.Warnf("Failed to create admin token: %v. Attempting to clean up infrastructure...", err)
			cleanup.Cleanup(t, terraformOptions, keyPath)
		}

		return nil, fmt.Errorf("failed to create admin token: %w", err)
	}

	rancherConfig.AdminToken = adminToken.Token

	var client *rancher.Client
	err = kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaults.TenMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		client, err = rancher.NewClient(rancherConfig.AdminToken, session)
		if err != nil {
			logrus.Warnf("Failed to create Rancher client: %v. Retrying...", err)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	client.RancherConfig.AdminToken = rancherConfig.AdminToken
	client.RancherConfig.AdminPassword = rancherConfig.AdminPassword
	client.RancherConfig.Host = host

	if !isUpgrade {
		err = pipeline.PostRancherInstall(client, client.RancherConfig.AdminPassword)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
