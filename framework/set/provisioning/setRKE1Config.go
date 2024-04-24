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
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	ec2 "github.com/rancher/tfp-automation/framework/set/provisioning/providers/ec2"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	clusterSync     = "rancher2_cluster_sync"
	nodeTemplate    = "rancher2_node_template"
	rancherNodePool = "rancher2_node_pool"

	backupConfig   = "backup_config"
	enabled        = "enabled"
	intervalHours  = "interval_hours"
	safeTimestamp  = "safe_timestamp"
	timeout        = "timeout"
	retention      = "retention"
	snapshot       = "snapshot"
	s3BackupConfig = "s3_backup_config"
	bucketName     = "bucket_name"

	hostnamePrefix     = "hostname_prefix"
	nodeTemplateID     = "node_template_id"
	controlPlane       = "control_plane"
	worker             = "worker"
	rancherNodePoolIDs = "node_pool_ids"
	stateConfirm       = "state_confirm"
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

	nodeTemplateBlock := rootBody.AppendNewBlock(resource, []string{nodeTemplate, nodeTemplate})
	nodeTemplateBlockBody := nodeTemplateBlock.Body()

	nodeTemplateBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.NodeTemplateName))

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

	if strings.Contains(psact, rancherBaseline) {
		newFile, rootBody = SetBaselinePSACT(newFile, rootBody)

		rootBody.AppendNewline()
	}

	clusterBlock := rootBody.AppendNewBlock(resource, []string{cluster, cluster})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + nodeTemplate + "]")},
	}

	if psact == rancherBaseline {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + nodeTemplate + "," +
				podSecurityAdmission + "." + podSecurityAdmission + "]")},
		}
	}

	clusterBlockBody.SetAttributeRaw(dependsOn, dependsOnTemp)
	clusterBlockBody.SetAttributeValue(resourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(defaultPodSecurityAdmission, cty.StringVal(psact))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(rkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue(kubernetesVersion, cty.StringVal(k8sVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock(network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(plugin, cty.StringVal(terraformConfig.NetworkPlugin))

	rootBody.AppendNewline()

	servicesBlock := rkeConfigBlockBody.AppendNewBlock(services, nil)
	servicesBlockBody := servicesBlock.Body()

	if terraformConfig.ETCDRKE1 != nil {
		etcdBlock := servicesBlockBody.AppendNewBlock(etcd, nil)
		etcdBlockBody := etcdBlock.Body()

		backupConfigBlock := etcdBlockBody.AppendNewBlock(backupConfig, nil)
		backupConfigBlockBody := backupConfigBlock.Body()

		backupConfigBlockBody.SetAttributeValue(enabled, cty.BoolVal(true))
		backupConfigBlockBody.SetAttributeValue(intervalHours, cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.IntervalHours))
		backupConfigBlockBody.SetAttributeValue(safeTimestamp, cty.BoolVal(terraformConfig.ETCDRKE1.BackupConfig.SafeTimestamp))
		backupConfigBlockBody.SetAttributeValue(timeout, cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.Timeout))

		if terraformConfig.Module == modules.EC2RKE1 && terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig != nil {
			s3ConfigBlock := backupConfigBlockBody.AppendNewBlock(s3BackupConfig, nil)
			s3ConfigBlockBody := s3ConfigBlock.Body()

			s3ConfigBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.AccessKey))
			s3ConfigBlockBody.SetAttributeValue(bucketName, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.BucketName))
			s3ConfigBlockBody.SetAttributeValue(endpoint, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Endpoint))
			s3ConfigBlockBody.SetAttributeValue(folder, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Folder))
			s3ConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Region))
			s3ConfigBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.SecretKey))
		}

		etcdBlockBody.SetAttributeValue(retention, cty.StringVal(terraformConfig.ETCDRKE1.Retention))
		etcdBlockBody.SetAttributeValue(snapshot, cty.BoolVal(false))
	}

	rootBody.AppendNewline()

	clusterSyncNodePoolIDs := ""

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolBlock := rootBody.AppendNewBlock(resource, []string{rancherNodePool, nodePool + poolNum})
		nodePoolBlockBody := nodePoolBlock.Body()

		dependsOnCluster := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + cluster + "." + cluster + "]")},
		}

		nodePoolBlockBody.SetAttributeRaw(dependsOn, dependsOnCluster)

		clusterID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(cluster + "." + cluster + ".id")},
		}

		nodePoolBlockBody.SetAttributeRaw(rancherClusterID, clusterID)
		nodePoolBlockBody.SetAttributeValue(resourceName, cty.StringVal(poolName+poolNum))
		nodePoolBlockBody.SetAttributeValue(hostnamePrefix, cty.StringVal(terraformConfig.HostnamePrefix+"-"+poolName))

		nodeTempID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(nodeTemplate + "." + nodeTemplate + ".id")},
		}

		nodePoolBlockBody.SetAttributeRaw(nodeTemplateID, nodeTempID)
		nodePoolBlockBody.SetAttributeValue(quantity, cty.NumberIntVal(pool.Quantity))
		nodePoolBlockBody.SetAttributeValue(controlPlane, cty.BoolVal(pool.Controlplane))
		nodePoolBlockBody.SetAttributeValue(etcd, cty.BoolVal(pool.Etcd))
		nodePoolBlockBody.SetAttributeValue(worker, cty.BoolVal(pool.Worker))

		rootBody.AppendNewline()

		if count != len(nodePools) {
			clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rancherNodePool + "." + nodePool + poolNum + ".id, "
		}

		if count == len(nodePools) {
			clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + rancherNodePool + "." + nodePool + poolNum + ".id"
		}
	}

	clusterSyncBlock := rootBody.AppendNewBlock(resource, []string{clusterSync, clusterSync})
	clusterSyncBlockBody := clusterSyncBlock.Body()

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(cluster + "." + cluster + ".id")},
	}

	clusterSyncBlockBody.SetAttributeRaw(rancherClusterID, clusterID)

	nodePoolIDs := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + clusterSyncNodePoolIDs + "]")},
	}

	clusterSyncBlockBody.SetAttributeRaw(rancherNodePoolIDs, nodePoolIDs)
	clusterSyncBlockBody.SetAttributeValue(stateConfirm, cty.NumberIntVal(2))

	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE1 configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
