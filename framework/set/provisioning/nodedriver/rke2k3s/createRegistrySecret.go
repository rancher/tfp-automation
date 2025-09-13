package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	clusterID    = "cluster_id"
	localCluster = "local"
	namespace    = "fleet-default"
	password     = "password"
	secretType   = "kubernetes.io/basic-auth"
	username     = "username"
)

// CreateRegistrySecret is a function that will set the airgap RKE2/K3s cluster configurations in the main.tf file.
func CreateRegistrySecret(terraformConfig *config.TerraformConfig, rootBody *hclwrite.Body) {
	secretBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.SecretV2, terraformConfig.ResourcePrefix})
	secretBlockBody := secretBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Rancher2 + "." + defaults.AdminUser)},
	}

	secretBlockBody.SetAttributeRaw(defaults.Provider, provider)

	secretBlockBody.SetAttributeValue(clusterID, cty.StringVal(localCluster))

	registrySecretName := terraformConfig.PrivateRegistries.AuthConfigSecretName + "-" + terraformConfig.ResourcePrefix

	secretBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(registrySecretName))
	secretBlockBody.SetAttributeValue(defaults.Namespace, cty.StringVal(namespace))
	secretBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(secretType))

	dataBlock := secretBlockBody.AppendNewBlock(defaults.Data+" =", nil)
	configBlockBody := dataBlock.Body()

	configBlockBody.SetAttributeValue(password, cty.StringVal(terraformConfig.PrivateRegistries.Password))
	configBlockBody.SetAttributeValue(username, cty.StringVal(terraformConfig.PrivateRegistries.Username))
}
