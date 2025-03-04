package rke1

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

func setNodePool(nodePools []config.Nodepool, count int, pool config.Nodepool, rootBody *hclwrite.Body,
	clusterSyncNodePoolIDs string, terraformConfig *config.TerraformConfig) error {
	poolNum := strconv.Itoa(count)

	_, err := resources.SetResourceNodepoolValidation(terraformConfig, pool, poolNum)
	if err != nil {
		return err
	}

	nodePoolBlock := rootBody.AppendNewBlock(defaults.Resource, []string{rancherNodePool, terraformConfig.ResourcePrefix + defaults.NodePool + poolNum})
	nodePoolBlockBody := nodePoolBlock.Body()

	dependsOnCluster := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + defaults.Cluster + "." + terraformConfig.ResourcePrefix + "]")},
	}

	nodePoolBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnCluster)

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Cluster + "." + terraformConfig.ResourcePrefix + ".id")},
	}

	nodePoolBlockBody.SetAttributeRaw(defaults.RancherClusterID, clusterID)
	nodePoolBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal("pool"+poolNum))
	nodePoolBlockBody.SetAttributeValue(hostnamePrefix, cty.StringVal(terraformConfig.ResourcePrefix+"-pool"+poolNum))

	nodeTempID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(nodeTemplate + "." + terraformConfig.ResourcePrefix + ".id")},
	}

	nodePoolBlockBody.SetAttributeRaw(nodeTemplateID, nodeTempID)
	nodePoolBlockBody.SetAttributeValue(defaults.Quantity, cty.NumberIntVal(pool.Quantity))
	nodePoolBlockBody.SetAttributeValue(controlPlane, cty.BoolVal(pool.Controlplane))
	nodePoolBlockBody.SetAttributeValue(defaults.Etcd, cty.BoolVal(pool.Etcd))
	nodePoolBlockBody.SetAttributeValue(worker, cty.BoolVal(pool.Worker))

	rootBody.AppendNewline()

	if count != len(nodePools) {
		clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rancherNodePool + "." + terraformConfig.ResourcePrefix + poolNum + ".id, "
	}

	if count == len(nodePools) {
		clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rancherNodePool + "." + terraformConfig.ResourcePrefix + poolNum + ".id"
	}

	return nil
}
