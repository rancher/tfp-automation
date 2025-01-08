package airgap

import (
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

// createRegistrySecret is a function that will set the airgap RKE2/K3s cluster configurations in the main.tf file.
func createRegistrySecret(terraformConfig *config.TerraformConfig, clusterName string, rootBody *hclwrite.Body) {
	secretBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.SecretV2, clusterName})
	secretBlockBody := secretBlock.Body()

	secretBlockBody.SetAttributeValue(clusterID, cty.StringVal(localCluster))
	secretBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.PrivateRegistries.AuthConfigSecretName))
	secretBlockBody.SetAttributeValue(defaults.Namespace, cty.StringVal(namespace))
	secretBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(secretType))

	dataBlock := secretBlockBody.AppendNewBlock(defaults.Data+" =", nil)
	configBlockBody := dataBlock.Body()

	configBlockBody.SetAttributeValue(password, cty.StringVal(terraformConfig.PrivateRegistries.Password))
	configBlockBody.SetAttributeValue(username, cty.StringVal(terraformConfig.PrivateRegistries.Username))
}
