package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// setClusterConfig is a function that will set the cluster configurations in the main.tf file.
func setClusterConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, psact, kubernetesVersion string) (*hclwrite.Body, error) {
	clusterBlock := rootBody.AppendNewBlock(general.Resource, []string{clusterV2, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(clusters.KubernetesVersion, cty.StringVal(kubernetesVersion))
	clusterBlockBody.SetAttributeValue(clusters.EnableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue(rancher2.DefaultPodSecurityAdmission, cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue(clusters.DefaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	return clusterBlockBody, nil
}
