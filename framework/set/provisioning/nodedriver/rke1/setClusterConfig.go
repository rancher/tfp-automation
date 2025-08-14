package rke1

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// setClusterConfig is a function that will set the cluster configurations in the main.tf file.
func setClusterConfig(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, psact string) (*hclwrite.Body, error) {
	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + terraformConfig.ResourcePrefix + "]")},
	}

	if psact == defaults.RancherBaseline {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + terraformConfig.ResourcePrefix + "," +
				defaults.PodSecurityAdmission + "." + terraformConfig.ResourcePrefix + "]")},
		}
	}

	clusterBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnTemp)
	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(defaults.DefaultPodSecurityAdmission, cty.StringVal(psact))

	return clusterBlockBody, nil
}
