package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/rancher/rancher/tests/v2/actions/pipeline"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/token"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// PostRancherSetup is a helper function that creates a Rancher client and accepts the EULA, if needed
func PostRancherSetup(t *testing.T, rancherConfig *rancher.Config, session *session.Session, host string, showToken, isAirgap bool) (*rancher.Client, error) {
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
	require.NoError(t, err)

	rancherConfig.AdminToken = adminToken.Token

	client, err := rancher.NewClient(rancherConfig.AdminToken, session)
	require.NoError(t, err)

	client.RancherConfig.AdminToken = rancherConfig.AdminToken
	client.RancherConfig.AdminPassword = rancherConfig.AdminPassword
	client.RancherConfig.Host = host

	err = pipeline.PostRancherInstall(client, client.RancherConfig.AdminPassword)
	require.NoError(t, err)

	// The FQDN needs to be set back to the Rancher host URL and not the internal FQDN
	if isAirgap {
		client.RancherConfig.Host = rancherConfig.Host
	}

	if showToken {
		logrus.Infof("Admin bearer token: %s", client.RancherConfig.AdminToken)
	}

	return client, nil
}
