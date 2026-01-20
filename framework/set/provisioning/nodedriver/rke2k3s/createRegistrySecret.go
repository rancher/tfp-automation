package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
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
	secretBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.SecretV2, terraformConfig.ResourcePrefix})
	secretBlockBody := secretBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(general.Rancher2 + "." + general.AdminUser)},
	}

	secretBlockBody.SetAttributeRaw(general.Provider, provider)

	secretBlockBody.SetAttributeValue(clusterID, cty.StringVal(localCluster))

	registrySecretName := terraformConfig.PrivateRegistries.AuthConfigSecretName + "-" + terraformConfig.ResourcePrefix

	secretBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(registrySecretName))
	secretBlockBody.SetAttributeValue(general.Namespace, cty.StringVal(namespace))
	secretBlockBody.SetAttributeValue(general.Type, cty.StringVal(secretType))

	dataBlock := secretBlockBody.AppendNewBlock(general.Data+" =", nil)
	configBlockBody := dataBlock.Body()

	configBlockBody.SetAttributeValue(password, cty.StringVal(terraformConfig.PrivateRegistries.Password))
	configBlockBody.SetAttributeValue(username, cty.StringVal(terraformConfig.PrivateRegistries.Username))
}
