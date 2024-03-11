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
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	ec2 "github.com/rancher/tfp-automation/framework/set/provisioning/providers/ec2"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetRKE1 is a function that will set the RKE1 configurations in the main.tf file.
func SetRKE1(clusterName, k8sVersion, psact string, nodePools []config.Nodepool, snapshots config.Snapshots, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig("rancher", rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig("terraform", terraformConfig)

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	newFile, rootBody := SetProvidersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	nodeTemplateBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_node_template", "rancher2_node_template"})
	nodeTemplateBlockBody := nodeTemplateBlock.Body()

	nodeTemplateBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.NodeTemplateName))

	switch {
	case terraformConfig.Module == azure_rke1:
		azure.SetAzureRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == ec2RKE1:
		ec2.SetEC2RKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == linodeRKE1:
		linode.SetLinodeRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == vsphereRKE1:
		vsphere.SetVsphereRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(psact, rancherBaseline) {
		newFile, rootBody = SetBaselinePSACT(newFile, rootBody)
		
		rootBody.AppendNewline()
	}

	clusterBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cluster", "rancher2_cluster"})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[rancher2_node_template.rancher2_node_template]")},
	}

	if psact == "rancher-baseline" {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[rancher2_node_template.rancher2_node_template," + 
			"rancher2_pod_security_admission_configuration_template.rancher2_pod_security_admission_configuration_template]")},
		}	
	}

	clusterBlockBody.SetAttributeRaw("depends_on", dependsOnTemp)
	clusterBlockBody.SetAttributeValue("name", cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue("default_pod_security_admission_configuration_template_name", cty.StringVal(psact))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock("rke_config", nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue("kubernetes_version", cty.StringVal(k8sVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock("network", nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue("plugin", cty.StringVal(terraformConfig.NetworkPlugin))

	rootBody.AppendNewline()

	servicesBlock := rkeConfigBlockBody.AppendNewBlock("services", nil)
	servicesBlockBody := servicesBlock.Body()

	if terraformConfig.ETCDRKE1 != nil {
		etcdBlock := servicesBlockBody.AppendNewBlock("etcd", nil)
		etcdBlockBody := etcdBlock.Body()

		backupConfigBlock := etcdBlockBody.AppendNewBlock("backup_config", nil)
		backupConfigBlockBody := backupConfigBlock.Body()

		backupConfigBlockBody.SetAttributeValue("enabled", cty.BoolVal(true))
		backupConfigBlockBody.SetAttributeValue("interval_hours", cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.IntervalHours))
		backupConfigBlockBody.SetAttributeValue("safe_timestamp", cty.BoolVal(terraformConfig.ETCDRKE1.BackupConfig.SafeTimestamp))
		backupConfigBlockBody.SetAttributeValue("timeout", cty.NumberIntVal(terraformConfig.ETCDRKE1.BackupConfig.Timeout))

		if terraformConfig.Module == ec2RKE1 && terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig != nil {
			s3ConfigBlock := backupConfigBlockBody.AppendNewBlock("s3_backup_config", nil)
			s3ConfigBlockBody := s3ConfigBlock.Body()

			s3ConfigBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.AccessKey))
			s3ConfigBlockBody.SetAttributeValue("bucket_name", cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.BucketName))
			s3ConfigBlockBody.SetAttributeValue("endpoint", cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Endpoint))
			s3ConfigBlockBody.SetAttributeValue("folder", cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Folder))
			s3ConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.Region))
			s3ConfigBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.ETCDRKE1.BackupConfig.S3BackupConfig.SecretKey))
		}

		etcdBlockBody.SetAttributeValue("retention", cty.StringVal(terraformConfig.ETCDRKE1.Retention))
		etcdBlockBody.SetAttributeValue("snapshot", cty.BoolVal(false))
	}

	rootBody.AppendNewline()

	clusterSyncNodePoolIDs := ""
	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_node_pool", "pool" + poolNum})
		nodePoolBlockBody := nodePoolBlock.Body()

		dependsOnCluster := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[rancher2_cluster.rancher2_cluster]")},
		}

		nodePoolBlockBody.SetAttributeRaw("depends_on", dependsOnCluster)

		clusterID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_cluster.rancher2_cluster.id")},
		}

		nodePoolBlockBody.SetAttributeRaw("cluster_id", clusterID)
		nodePoolBlockBody.SetAttributeValue("name", cty.StringVal("pool"+poolNum))
		nodePoolBlockBody.SetAttributeValue("hostname_prefix", cty.StringVal(terraformConfig.HostnamePrefix+"-pool"+poolNum+"-"))

		nodeTempID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_node_template.rancher2_node_template.id")},
		}

		nodePoolBlockBody.SetAttributeRaw("node_template_id", nodeTempID)
		nodePoolBlockBody.SetAttributeValue("quantity", cty.NumberIntVal(pool.Quantity))
		nodePoolBlockBody.SetAttributeValue("control_plane", cty.BoolVal(pool.Controlplane))
		nodePoolBlockBody.SetAttributeValue("etcd", cty.BoolVal(pool.Etcd))
		nodePoolBlockBody.SetAttributeValue("worker", cty.BoolVal(pool.Worker))

		rootBody.AppendNewline()

		if count != len(nodePools) {
			clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + "rancher2_node_pool.pool" + poolNum + ".id, "
		}
		if count == len(nodePools) {
			clusterSyncNodePoolIDs = clusterSyncNodePoolIDs + "rancher2_node_pool.pool" + poolNum + ".id"
		}

		count++
	}

	clusterSyncBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cluster_sync", "rancher2_cluster_sync"})
	clusterSyncBlockBody := clusterSyncBlock.Body()

	clusterID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_cluster.rancher2_cluster.id")},
	}

	clusterSyncBlockBody.SetAttributeRaw("cluster_id", clusterID)

	nodePoolIDs := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + clusterSyncNodePoolIDs + "]")},
	}

	clusterSyncBlockBody.SetAttributeRaw("node_pool_ids", nodePoolIDs)
	clusterSyncBlockBody.SetAttributeValue("state_confirm", cty.NumberIntVal(2))

	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE1 configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
