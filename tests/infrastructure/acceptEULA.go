package infrastructure

import (
	"os"
	"testing"

	"github.com/rancher/rancher/tests/v2/actions/pipeline"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/token"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/tests/extensions/provisioning"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// AcceptEULA accepts the EULA for the Rancher server post installation
func AcceptEULA(t *testing.T, session *session.Session, host string) {
	cattleConfig := shepherdConfig.LoadConfigFromFile(os.Getenv(shepherdConfig.ConfigEnvironmentKey))
	configMap, err := provisioning.UniquifyTerraform([]map[string]any{cattleConfig})
	require.NoError(t, err)

	cattleConfig = configMap[0]
	rancherConfig, _, _ := config.LoadTFPConfigs(cattleConfig)

	adminUser := &management.User{
		Username: "admin",
		Password: rancherConfig.AdminPassword,
	}

	userToken, err := token.GenerateUserToken(adminUser, rancherConfig.Host)
	require.NoError(t, err)

	rancherConfig.AdminToken = userToken.Token

	client, err := rancher.NewClient(rancherConfig.AdminToken, session)
	require.NoError(t, err)

	client.RancherConfig.AdminToken = rancherConfig.AdminToken
	client.RancherConfig.AdminPassword = rancherConfig.AdminPassword
	client.RancherConfig.Host = host

	_, err = operations.ReplaceValue([]string{"rancher", "adminToken"}, rancherConfig.AdminToken, configMap[0])
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"rancher", "adminPassword"}, rancherConfig.AdminPassword, configMap[0])
	require.NoError(t, err)

	_, err = operations.ReplaceValue([]string{"rancher", "host"}, rancherConfig.Host, configMap[0])
	require.NoError(t, err)

	err = pipeline.PostRancherInstall(client, client.RancherConfig.AdminPassword)
	require.NoError(t, err)

	logrus.Infof("Admin bearer token: %s", client.RancherConfig.AdminToken)
}
