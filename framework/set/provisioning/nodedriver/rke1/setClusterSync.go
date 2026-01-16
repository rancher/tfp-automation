package rke1

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

func setClusterSync(rootBody *hclwrite.Body, clusterSyncNodePoolIDs string, clusterName string) error {
	clusterSyncBlock := rootBody.AppendNewBlock(general.Resource, []string{clusterSync, clusterName})
	clusterSyncBlockBody := clusterSyncBlock.Body()

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancher2.Cluster + "." + clusterName + ".id")},
	}

	clusterSyncBlockBody.SetAttributeRaw(clusters.RancherClusterID, clusterID)

	nodePoolIDs := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + clusterSyncNodePoolIDs + "]")},
	}

	clusterSyncBlockBody.SetAttributeRaw(rancherNodePoolIDs, nodePoolIDs)
	clusterSyncBlockBody.SetAttributeValue(stateConfirm, cty.NumberIntVal(2))

	rootBody.AppendNewline()

	return nil
}
