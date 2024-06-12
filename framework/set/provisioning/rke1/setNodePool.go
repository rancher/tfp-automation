package rke1

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/zclconf/go-cty/cty"
)

func setNodePool(nodePools []config.Nodepool, count int, pool config.Nodepool, rootBody *hclwrite.Body,
	clusterSyncNodePoolIDs, poolName string, terraformConfig *config.TerraformConfig) error {
	poolNum := strconv.Itoa(count)

	_, err := resources.SetResourceNodepoolValidation(pool, poolNum)
	if err != nil {
		return err
	}

	nodePoolBlock := rootBody.AppendNewBlock(defaults.Resource, []string{rancherNodePool, defaults.NodePool + poolNum})
	nodePoolBlockBody := nodePoolBlock.Body()

	dependsOnCluster := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + defaults.Cluster + "." + defaults.Cluster + "]")},
	}

	nodePoolBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnCluster)

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Cluster + "." + defaults.Cluster + ".id")},
	}

	nodePoolBlockBody.SetAttributeRaw(defaults.RancherClusterID, clusterID)
	nodePoolBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(poolName+poolNum))
	nodePoolBlockBody.SetAttributeValue(hostnamePrefix, cty.StringVal(terraformConfig.HostnamePrefix+"-"+poolName+poolNum))

	nodeTempID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(nodeTemplate + "." + nodeTemplate + ".id")},
	}

	nodePoolBlockBody.SetAttributeRaw(nodeTemplateID, nodeTempID)
	nodePoolBlockBody.SetAttributeValue(defaults.Quantity, cty.NumberIntVal(pool.Quantity))
	nodePoolBlockBody.SetAttributeValue(controlPlane, cty.BoolVal(pool.Controlplane))
	nodePoolBlockBody.SetAttributeValue(defaults.Etcd, cty.BoolVal(pool.Etcd))
	nodePoolBlockBody.SetAttributeValue(worker, cty.BoolVal(pool.Worker))

	rootBody.AppendNewline()

	if count != len(nodePools) {
		clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rancherNodePool + "." + defaults.NodePool + poolNum + ".id, "
	}

	if count == len(nodePools) {
		clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rancherNodePool + "." + defaults.NodePool + poolNum + ".id"
	}

	return nil
}
