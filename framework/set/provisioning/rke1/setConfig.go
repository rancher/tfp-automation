package rke1

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	ec2 "github.com/rancher/tfp-automation/framework/set/provisioning/providers/ec2"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/resources"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	clusterSync     = "rancher2_cluster_sync"
	nodeTemplate    = "rancher2_node_template"
	rancherNodePool = "rancher2_node_pool"

	backupConfig   = "backup_config"
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

	newFile, rootBody := resources.SetProvidersAndUsersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	nodeTemplateBlock := rootBody.AppendNewBlock(defaults.Resource, []string{nodeTemplate, nodeTemplate})
	nodeTemplateBlockBody := nodeTemplateBlock.Body()

	nodeTemplateBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.NodeTemplateName))

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

	if strings.Contains(psact, defaults.RancherBaseline) {
		newFile, rootBody = resources.SetBaselinePSACT(newFile, rootBody)

		rootBody.AppendNewline()
	}

	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, defaults.Cluster})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + nodeTemplate + "]")},
	}

	if psact == defaults.RancherBaseline {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + nodeTemplate + "," +
				defaults.PodSecurityAdmission + "." + defaults.PodSecurityAdmission + "]")},
		}
	}

	clusterBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnTemp)
	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(defaults.DefaultPodSecurityAdmission, cty.StringVal(psact))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(k8sVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(defaults.Plugin, cty.StringVal(terraformConfig.NetworkPlugin))

	rootBody.AppendNewline()

	servicesBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Services, nil)
	servicesBlockBody := servicesBlock.Body()

	if terraformConfig.ETCDRKE1 != nil {
		setEtcdConfig(servicesBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	clusterSyncNodePoolIDs := ""

	for count, pool := range nodePools {
		setNodePool(nodePools, count, pool, rootBody, clusterSyncNodePoolIDs, poolName, terraformConfig)
	}

	setClusterSync(rootBody, clusterSyncNodePoolIDs)

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE1 configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
