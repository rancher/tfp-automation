package rke1

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

// setClusterConfig is a function that will set the cluster configurations in the main.tf file.
func setClusterConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, psact string) (*hclwrite.Body, error) {
	clusterBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.Cluster, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + terraformConfig.ResourcePrefix + "]")},
	}

	if psact == clusters.RancherBaseline {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + terraformConfig.ResourcePrefix + "," +
				rancher2.PodSecurityAdmission + "." + terraformConfig.ResourcePrefix + "]")},
		}
	}

	clusterBlockBody.SetAttributeRaw(general.DependsOn, dependsOnTemp)
	clusterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(rancher2.DefaultPodSecurityAdmission, cty.StringVal(psact))

	return clusterBlockBody, nil
}
