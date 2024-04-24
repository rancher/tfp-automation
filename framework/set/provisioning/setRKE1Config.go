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
	rke1Block "github.com/rancher/tfp-automation/defaults/resourceblocks/rke1"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	ec2 "github.com/rancher/tfp-automation/framework/set/provisioning/providers/ec2"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetRKE1 is a function that will set the RKE1 configurations in the main.tf file.
func SetRKE1(clusterName, poolName, k8sVersion, psact string, nodePools []config.Nodepool, snapshots config.Snapshots, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(configs.Terraform, terraformConfig)

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	newFile, rootBody := SetProvidersAndUsersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	nodeTemplateBlock := rootBody.AppendNewBlock(blocks.Resource, []string{rke1Block.NodeTemplate, rke1Block.NodeTemplate})
	nodeTemplateBlockBody := nodeTemplateBlock.Body()

	nodeTemplateBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(terraformConfig.NodeTemplateName))

	switch {
	case terraformConfig.Module == modules.AzureRKE1:
		azure.SetAzureRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.EC2RKE1:
		ec2.SetEC2RKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.LinodeRKE1:
		linode.SetLinodeRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.VsphereRKE1:
		vsphere.SetVsphereRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(psact, psactBlock.RancherBaseline) {
		newFile, rootBody = SetBaselinePSACT(newFile, rootBody)

		rootBody.AppendNewline()
	}

	clusterBlock := rootBody.AppendNewBlock(blocks.Resource, []string{blocks.Cluster, blocks.Cluster})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + rke1Block.NodeTemplate + "." + rke1Block.NodeTemplate + "]")},
	}

	if psact == psactBlock.RancherBaseline {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + rke1Block.NodeTemplate + "." + rke1Block.NodeTemplate + "," +
				blocks.PodSecurityAdmission + "." + blocks.PodSecurityAdmission + "]")},
		}
	}

	clusterBlockBody.SetAttributeRaw(blocks.DependsOn, dependsOnTemp)
	clusterBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(blocks.DefaultPodSecurityAdmission, cty.StringVal(psact))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(blocks.RKEConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue(blocks.KubernetesVersion, cty.StringVal(k8sVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock(blocks.Network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(blocks.Plugin, cty.StringVal(terraformConfig.NetworkPlugin))

	rootBody.AppendNewline()

	servicesBlock := rkeConfigBlockBody.AppendNewBlock(blocks.Services, nil)
	servicesBlockBody := servicesBlock.Body()

	if terraformConfig.ETCDRKE1 != nil {
		etcdBlock := servicesBlockBody.AppendNewBlock(blocks.Etcd, nil)
		etcdBlockBody := etcdBlock.Body()

		backupConfigBlock := etcdBlockBody.AppendNewBlock(rke1Block.BackupConfig, nil)
		backupConfigBlockBody := backupConfigBlock.Body()

		backupConfigBlockBody.SetAttributeValue(rke1Block.Enabled, cty.BoolVal(true))
		backupConfigBlockBody.SetAttributeValue(rke1Block.IntervalHours, cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.IntervalHours))
		backupConfigBlockBody.SetAttributeValue(rke1Block.SafeTimestamp, cty.BoolVal(terraformConfig.ETCDRKE1.BackupConfig.SafeTimestamp))
		backupConfigBlockBody.SetAttributeValue(rke1Block.Timeout, cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.Timeout))

		if terraformConfig.Module == modules.EC2RKE1 && terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig != nil {
			s3ConfigBlock := backupConfigBlockBody.AppendNewBlock(rke1Block.S3BackupConfig, nil)
			s3ConfigBlockBody := s3ConfigBlock.Body()

			s3ConfigBlockBody.SetAttributeValue(blocks.AccessKey, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.AccessKey))
			s3ConfigBlockBody.SetAttributeValue(rke1Block.BucketName, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.BucketName))
			s3ConfigBlockBody.SetAttributeValue(blocks.Endpoint, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Endpoint))
			s3ConfigBlockBody.SetAttributeValue(blocks.Folder, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Folder))
			s3ConfigBlockBody.SetAttributeValue(blocks.Region, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Region))
			s3ConfigBlockBody.SetAttributeValue(blocks.SecretKey, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.SecretKey))
		}

		etcdBlockBody.SetAttributeValue(rke1Block.Retention, cty.StringVal(terraformConfig.ETCDRKE1.Retention))
		etcdBlockBody.SetAttributeValue(rke1Block.Snapshot, cty.BoolVal(false))
	}

	rootBody.AppendNewline()

	clusterSyncNodePoolIDs := ""

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolBlock := rootBody.AppendNewBlock(blocks.Resource, []string{rke1Block.NodePool, blocks.Pool + poolNum})
		nodePoolBlockBody := nodePoolBlock.Body()

		dependsOnCluster := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + blocks.Cluster + "." + blocks.Cluster + "]")},
		}

		nodePoolBlockBody.SetAttributeRaw(blocks.DependsOn, dependsOnCluster)

		clusterID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(blocks.Cluster + "." + blocks.Cluster + ".id")},
		}

		nodePoolBlockBody.SetAttributeRaw(blocks.ClusterID, clusterID)
		nodePoolBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(poolName+poolNum))
		nodePoolBlockBody.SetAttributeValue(rke1Block.HostnamePrefix, cty.StringVal(terraformConfig.HostnamePrefix+"-"+poolName))

		nodeTempID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(rke1Block.NodeTemplate + "." + rke1Block.NodeTemplate + ".id")},
		}

		nodePoolBlockBody.SetAttributeRaw(rke1Block.NodeTemplateID, nodeTempID)
		nodePoolBlockBody.SetAttributeValue(blocks.Quantity, cty.NumberIntVal(pool.Quantity))
		nodePoolBlockBody.SetAttributeValue(rke1Block.ControlPlane, cty.BoolVal(pool.Controlplane))
		nodePoolBlockBody.SetAttributeValue(blocks.Etcd, cty.BoolVal(pool.Etcd))
		nodePoolBlockBody.SetAttributeValue(rke1Block.Worker, cty.BoolVal(pool.Worker))

		rootBody.AppendNewline()

		if count != len(nodePools) {
			clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rke1Block.NodePool + "." + blocks.Pool + poolNum + ".id, "
		}

		if count == len(nodePools) {
			clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rke1Block.NodePool + "." + blocks.Pool + poolNum + ".id"
		}
	}

	clusterSyncBlock := rootBody.AppendNewBlock(blocks.Resource, []string{rke1Block.ClusterSync, rke1Block.ClusterSync})
	clusterSyncBlockBody := clusterSyncBlock.Body()

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(blocks.Cluster + "." + blocks.Cluster + ".id")},
	}

	clusterSyncBlockBody.SetAttributeRaw(blocks.ClusterID, clusterID)

	nodePoolIDs := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + clusterSyncNodePoolIDs + "]")},
	}

	clusterSyncBlockBody.SetAttributeRaw(rke1Block.NodePoolIDs, nodePoolIDs)
	clusterSyncBlockBody.SetAttributeValue(rke1Block.StateConfirm, cty.NumberIntVal(2))

	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE1 configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
