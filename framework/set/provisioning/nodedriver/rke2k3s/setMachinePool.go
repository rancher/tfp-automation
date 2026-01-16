package rke2k3s

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

func setMachinePool(terraformConfig *config.TerraformConfig, count int, pool config.Nodepool, rkeConfigBlockBody *hclwrite.Body) error {
	poolNum := strconv.Itoa(count)

	_, err := resources.SetResourceNodepoolValidation(terraformConfig, pool, poolNum)
	if err != nil {
		return err
	}

	machinePoolsBlock := rkeConfigBlockBody.AppendNewBlock(clusters.MachinePools, nil)
	machinePoolsBlockBody := machinePoolsBlock.Body()

	machinePoolsBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+poolNum))

	cloudCredSecretName := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancher2.CloudCredential + "." + terraformConfig.ResourcePrefix + ".id")},
	}

	machinePoolsBlockBody.SetAttributeRaw(cloudCredentialSecretName, cloudCredSecretName)
	machinePoolsBlockBody.SetAttributeValue(controlPlaneRole, cty.BoolVal(pool.Controlplane))
	machinePoolsBlockBody.SetAttributeValue(etcdRole, cty.BoolVal(pool.Etcd))
	machinePoolsBlockBody.SetAttributeValue(workerRole, cty.BoolVal(pool.Worker))
	machinePoolsBlockBody.SetAttributeValue(clusters.Quantity, cty.NumberIntVal(pool.Quantity))

	machineConfigBlock := machinePoolsBlockBody.AppendNewBlock(clusters.MachineConfig, nil)
	machineConfigBlockBody := machineConfigBlock.Body()

	kind := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(machineConfigV2 + "." + terraformConfig.ResourcePrefix + ".kind")},
	}

	machineConfigBlockBody.SetAttributeRaw(general.ResourceKind, kind)

	name := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(machineConfigV2 + "." + terraformConfig.ResourcePrefix + ".name")},
	}

	machineConfigBlockBody.SetAttributeRaw(general.ResourceName, name)

	return nil
}
