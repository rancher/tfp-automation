package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setClusterConfig is a function that will set the cluster configurations in the main.tf file.
func setClusterConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, provider hclwrite.Tokens, psact,
	kubernetesVersion string) (*hclwrite.Body, error) {
	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{clusterV2, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeRaw(defaults.Provider, provider)
	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(kubernetesVersion))
	clusterBlockBody.SetAttributeValue(defaults.EnableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue(defaults.DefaultPodSecurityAdmission, cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue(defaults.DefaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	return clusterBlockBody, nil
}
