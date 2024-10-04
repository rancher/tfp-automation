package rke1

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

func setClusterSync(rootBody *hclwrite.Body, clusterSyncNodePoolIDs string) error {
	clusterSyncBlock := rootBody.AppendNewBlock(defaults.Resource, []string{clusterSync, clusterSync})
	clusterSyncBlockBody := clusterSyncBlock.Body()

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Cluster + "." + defaults.Cluster + ".id")},
	}

	clusterSyncBlockBody.SetAttributeRaw(defaults.RancherClusterID, clusterID)

	nodePoolIDs := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + clusterSyncNodePoolIDs + "]")},
	}

	clusterSyncBlockBody.SetAttributeRaw(rancherNodePoolIDs, nodePoolIDs)
	clusterSyncBlockBody.SetAttributeValue(stateConfirm, cty.NumberIntVal(2))

	rootBody.AppendNewline()

	return nil
}
