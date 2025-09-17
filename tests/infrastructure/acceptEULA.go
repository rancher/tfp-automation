package infrastructure

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/tests/actions/pipeline"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/stretchr/testify/require"
)

// PostRancherSetup is a helper function that creates a Rancher client and accepts the EULA, if needed
func PostRancherSetup(t *testing.T, terraformOptions *terraform.Options, rancherConfig *rancher.Config, session *session.Session, host,
	keyPath string, isAirgap, isUpgrade bool) (*rancher.Client, error) {
	adminToken, err := CreateAdminToken(t, terraformOptions, rancherConfig, keyPath)
	if err != nil && *rancherConfig.Cleanup {
		cleanup.Cleanup(t, terraformOptions, keyPath)
	}

	rancherConfig.AdminToken = adminToken.Token

	client, err := rancher.NewClient(rancherConfig.AdminToken, session)
	require.NoError(t, err)

	client.RancherConfig.AdminToken = rancherConfig.AdminToken
	client.RancherConfig.AdminPassword = rancherConfig.AdminPassword
	client.RancherConfig.Host = host

	if !isUpgrade {
		err = pipeline.PostRancherInstall(client, client.RancherConfig.AdminPassword)
		require.NoError(t, err)
	}

	// The FQDN needs to be set back to the Rancher host URL and not the internal FQDN
	if isAirgap {
		client.RancherConfig.Host = rancherConfig.Host
	}

	return client, nil
}
