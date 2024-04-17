package provisioning

import (
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
	blocks "github.com/rancher/tfp-automation/defaults/resourceblocks"
	psactBlock "github.com/rancher/tfp-automation/defaults/resourceblocks/psact"
	v2Block "github.com/rancher/tfp-automation/defaults/resourceblocks/rke2k3s"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	ec2 "github.com/rancher/tfp-automation/framework/set/provisioning/providers/ec2"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetRKE2K3s is a function that will set the RKE2/K3S configurations in the main.tf file.
func SetRKE2K3s(clusterName, poolName, k8sVersion, psact string, nodePools []config.Nodepool, snapshots config.Snapshots, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	newFile, rootBody := SetProvidersAndUsersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	switch {
	case terraformConfig.Module == modules.AzureRKE2 || terraformConfig.Module == modules.AzureK3s:
		azure.SetAzureRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == modules.EC2RKE2 || terraformConfig.Module == modules.EC2K3s:
		ec2.SetEC2RKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == modules.LinodeRKE2 || terraformConfig.Module == modules.LinodeK3s:
		linode.SetLinodeRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == modules.VsphereRKE2 || terraformConfig.Module == modules.VsphereK3s:
		vsphere.SetVsphereRKE2K3SProvider(rootBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(psact, psactBlock.RancherBaseline) {
		newFile, rootBody = SetBaselinePSACT(newFile, rootBody)

		rootBody.AppendNewline()
	}

	machineConfigBlock := rootBody.AppendNewBlock(blocks.Resource, []string{v2Block.MachineConfigV2, v2Block.MachineConfigV2})
	machineConfigBlockBody := machineConfigBlock.Body()

	if psact == psactBlock.RancherBaseline {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + blocks.PodSecurityAdmission + "." +
				blocks.PodSecurityAdmission + "]")},
		}

		machineConfigBlockBody.SetAttributeRaw(blocks.DependsOn, dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue(blocks.GenerateName, cty.StringVal(terraformConfig.MachineConfigName))

	switch {
	case terraformConfig.Module == modules.AzureRKE2 || terraformConfig.Module == modules.AzureK3s:
		azure.SetAzureRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == modules.EC2RKE2 || terraformConfig.Module == modules.EC2K3s:
		ec2.SetEC2RKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == modules.LinodeRKE2 || terraformConfig.Module == modules.LinodeK3s:
		linode.SetLinodeRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == modules.VsphereRKE2 || terraformConfig.Module == modules.VsphereK3s:
		vsphere.SetVsphereRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(blocks.Resource, []string{v2Block.ClusterV2, v2Block.ClusterV2})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(blocks.KubernetesVersion, cty.StringVal(k8sVersion))
	clusterBlockBody.SetAttributeValue(blocks.EnableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue(blocks.DefaultPodSecurityAdmission, cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue(blocks.DefaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(blocks.RKEConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		machinePoolsBlock := rkeConfigBlockBody.AppendNewBlock(blocks.MachinePools, nil)
		machinePoolsBlockBody := machinePoolsBlock.Body()

		machinePoolsBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(poolName+poolNum))

		cloudCredSecretName := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(blocks.CloudCredential + "." + blocks.CloudCredential + ".id")},
		}

		machinePoolsBlockBody.SetAttributeRaw(v2Block.CloudCredentialSecretName, cloudCredSecretName)
		machinePoolsBlockBody.SetAttributeValue(v2Block.ControlPlaneRole, cty.BoolVal(pool.Controlplane))
		machinePoolsBlockBody.SetAttributeValue(v2Block.EtcdRole, cty.BoolVal(pool.Etcd))
		machinePoolsBlockBody.SetAttributeValue(v2Block.WorkerRole, cty.BoolVal(pool.Worker))
		machinePoolsBlockBody.SetAttributeValue(blocks.Quantity, cty.NumberIntVal(pool.Quantity))

		machineConfigBlock := machinePoolsBlockBody.AppendNewBlock(blocks.MachineConfig, nil)
		machineConfigBlockBody := machineConfigBlock.Body()

		kind := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(v2Block.MachineConfigV2 + "." + v2Block.MachineConfigV2 + ".kind")},
		}

		machineConfigBlockBody.SetAttributeRaw(blocks.ResourceKind, kind)

		name := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(v2Block.MachineConfigV2 + "." + v2Block.MachineConfigV2 + ".name")},
		}

		machineConfigBlockBody.SetAttributeRaw(blocks.ResourceName, name)
	}

	upgradeStrategyBlock := rkeConfigBlockBody.AppendNewBlock(v2Block.UpgradeStrategy, nil)
	upgradeStrategyBlockBody := upgradeStrategyBlock.Body()

	upgradeStrategyBlockBody.SetAttributeValue(v2Block.ControlPlaneConcurrency, cty.StringVal(("10%")))
	upgradeStrategyBlockBody.SetAttributeValue(v2Block.WorkerConcurrency, cty.StringVal(("10%")))

	if terraformConfig.ETCD != nil {
		snapshotBlock := rkeConfigBlockBody.AppendNewBlock(blocks.Etcd, nil)
		snapshotBlockBody := snapshotBlock.Body()

		snapshotBlockBody.SetAttributeValue(v2Block.DisableSnapshots, cty.BoolVal(terraformConfig.ETCD.DisableSnapshots))
		snapshotBlockBody.SetAttributeValue(v2Block.SnapshotScheduleCron, cty.StringVal(terraformConfig.ETCD.SnapshotScheduleCron))
		snapshotBlockBody.SetAttributeValue(v2Block.SnapshotRetention, cty.NumberIntVal(int64(terraformConfig.ETCD.SnapshotRetention)))

		if strings.Contains(terraformConfig.Module, modules.EC2) && terraformConfig.ETCD.S3 != nil {
			s3ConfigBlock := snapshotBlockBody.AppendNewBlock(v2Block.S3Config, nil)
			s3ConfigBlockBody := s3ConfigBlock.Body()

			cloudCredSecretName := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(blocks.CloudCredential + "." + blocks.CloudCredential + ".id")},
			}

			s3ConfigBlockBody.SetAttributeValue(v2Block.Bucket, cty.StringVal(terraformConfig.ETCD.S3.Bucket))
			s3ConfigBlockBody.SetAttributeValue(blocks.Endpoint, cty.StringVal(terraformConfig.ETCD.S3.Endpoint))
			s3ConfigBlockBody.SetAttributeRaw(v2Block.CloudCredentialName, cloudCredSecretName)
			s3ConfigBlockBody.SetAttributeValue(v2Block.EndpointCA, cty.StringVal(terraformConfig.ETCD.S3.EndpointCA))
			s3ConfigBlockBody.SetAttributeValue(blocks.Folder, cty.StringVal(terraformConfig.ETCD.S3.Folder))
			s3ConfigBlockBody.SetAttributeValue(blocks.Region, cty.StringVal(terraformConfig.ETCD.S3.Region))
			s3ConfigBlockBody.SetAttributeValue(v2Block.SkipSSLVerify, cty.BoolVal(terraformConfig.ETCD.S3.SkipSSLVerify))
		}
	}

	if snapshots.CreateSnapshot {
		setCreateRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody)
	}

	if snapshots.RestoreSnapshot {
		setRestoreRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody, snapshots)
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
