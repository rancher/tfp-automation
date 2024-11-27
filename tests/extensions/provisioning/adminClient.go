package provisioning

import (
	"testing"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/stretchr/testify/require"
)

// FetchAdminClient will return the admin client for the cluster.
func FetchAdminClient(t *testing.T, client *rancher.Client) (*rancher.Client, error) {
	client, err := client.ReLogin()
	require.NoError(t, err)

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	require.NoError(t, err)

	return adminClient, err
}
