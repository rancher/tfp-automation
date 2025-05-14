package infrastructure

import (
	"os"
	"testing"

	"github.com/rancher/rancher/tests/v2/actions/pipeline"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// AcceptEULA accepts the EULA for the Rancher server post installation
func AcceptEULA(t *testing.T, session *session.Session, host string, showToken, isAirgap bool) (*rancher.Client, error) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	rancherConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	adminToken, err := token.GenerateUserToken(adminUser, rancherConfig.Host)
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
