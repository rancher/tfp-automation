package rke2k3s

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/zclconf/go-cty/cty"
)

func setMachinePool(machinePools []config.Nodepool, count int, pool config.Nodepool, rkeConfigBlockBody *hclwrite.Body,
	poolName string) error {
	poolNum := strconv.Itoa(count)

	_, err := resources.SetResourceNodepoolValidation(pool, poolNum)
	if err != nil {
		return err
	}

	machinePoolsBlock := rkeConfigBlockBody.AppendNewBlock(defaults.MachinePools, nil)
	machinePoolsBlockBody := machinePoolsBlock.Body()

	machinePoolsBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(poolName+poolNum))

	cloudCredSecretName := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.CloudCredential + "." + defaults.CloudCredential + ".id")},
	}

	machinePoolsBlockBody.SetAttributeRaw(cloudCredentialSecretName, cloudCredSecretName)
	machinePoolsBlockBody.SetAttributeValue(controlPlaneRole, cty.BoolVal(pool.Controlplane))
	machinePoolsBlockBody.SetAttributeValue(etcdRole, cty.BoolVal(pool.Etcd))
	machinePoolsBlockBody.SetAttributeValue(workerRole, cty.BoolVal(pool.Worker))
	machinePoolsBlockBody.SetAttributeValue(defaults.Quantity, cty.NumberIntVal(pool.Quantity))

	machineConfigBlock := machinePoolsBlockBody.AppendNewBlock(defaults.MachineConfig, nil)
	machineConfigBlockBody := machineConfigBlock.Body()

	kind := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(machineConfigV2 + "." + machineConfigV2 + ".kind")},
	}

	machineConfigBlockBody.SetAttributeRaw(defaults.ResourceKind, kind)

	name := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(machineConfigV2 + "." + machineConfigV2 + ".name")},
	}

	machineConfigBlockBody.SetAttributeRaw(defaults.ResourceName, name)

	return nil
}
