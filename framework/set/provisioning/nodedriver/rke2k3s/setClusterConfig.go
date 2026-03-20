package rke2k3s

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// setClusterConfig is a function that will set the cluster configurations in the main.tf file.
func setClusterConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (*hclwrite.Body, error) {
	clusterBlock := rootBody.AppendNewBlock(general.Resource, []string{clusterV2, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(clusters.KubernetesVersion, cty.StringVal(terratestConfig.KubernetesVersion))
	clusterBlockBody.SetAttributeValue(clusters.EnableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))

	// If PSACT value contains "baseline", we specifically create a unique baseline PSACT for each cluster. So we
	// must specify that here.
	if strings.Contains(terratestConfig.PSACT, "baseline") {
		clusterBlockBody.SetAttributeValue(rancher2.DefaultPodSecurityAdmission, cty.StringVal(terratestConfig.PSACT+"-"+terraformConfig.ResourcePrefix))
	} else {
		clusterBlockBody.SetAttributeValue(rancher2.DefaultPodSecurityAdmission, cty.StringVal(terratestConfig.PSACT))
	}

	clusterBlockBody.SetAttributeValue(clusters.DefaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	return clusterBlockBody, nil
}
