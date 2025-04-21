package rke1

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	v2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	aws "github.com/rancher/tfp-automation/framework/set/provisioning/providers/aws"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	harvester "github.com/rancher/tfp-automation/framework/set/provisioning/providers/harvester"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	clusterSync     = "rancher2_cluster_sync"
	nodeTemplate    = "rancher2_node_template"
	rancherNodePool = "rancher2_node_pool"

	backupConfig            = "backup_config"
	intervalHours           = "interval_hours"
	safeTimestamp           = "safe_timestamp"
	timeout                 = "timeout"
	retention               = "retention"
	snapshot                = "snapshot"
	s3BackupConfig          = "s3_backup_config"
	bucketName              = "bucket_name"
	privateRegistryURL      = "url"
	privateRegistryUsername = "user"
	privateRegistryPassword = "password"

	hostnamePrefix     = "hostname_prefix"
	nodeTemplateID     = "node_template_id"
	controlPlane       = "control_plane"
	worker             = "worker"
	rancherNodePoolIDs = "node_pool_ids"
	stateConfirm       = "state_confirm"
	project            = "project"
)

// SetRKE1 is a function that will set the RKE1 configurations in the main.tf file.
func SetRKE1(terraformConfig *config.TerraformConfig, k8sVersion, psact string, nodePools []config.Nodepool, snapshots config.Snapshots,
	newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, rbacRole config.Role) (*hclwrite.File, *os.File, error) {

	nodeTemplateBlock := rootBody.AppendNewBlock(defaults.Resource, []string{nodeTemplate, terraformConfig.ResourcePrefix})
	nodeTemplateBlockBody := nodeTemplateBlock.Body()

	nodeTemplateBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	if terraformConfig.PrivateRegistries != nil {
		nodeTemplateBlockBody.SetAttributeValue(defaults.EngineInsecureRegistry, cty.ListVal([]cty.Value{
			cty.StringVal(terraformConfig.PrivateRegistries.URL),
		}))
	}

	switch {
	case terraformConfig.Module == modules.EC2RKE1:
		aws.SetAWSRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.AzureRKE1:
		azure.SetAzureRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.LinodeRKE1:
		linode.SetLinodeRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.HarvesterRKE1:
		harvester.SetHarvesterCredentialProvider(rootBody, terraformConfig)
		harvester.SetHarvesterRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case terraformConfig.Module == modules.VsphereRKE1:
		vsphere.SetVsphereRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(psact, defaults.RancherBaseline) {
		newFile, rootBody = resources.SetBaselinePSACT(newFile, rootBody, terraformConfig.ResourcePrefix)

		rootBody.AppendNewline()
	}

	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	dependsOnTemp := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + terraformConfig.ResourcePrefix + "]")},
	}

	if psact == defaults.RancherBaseline {
		dependsOnTemp = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + nodeTemplate + "." + terraformConfig.ResourcePrefix + "," +
				defaults.PodSecurityAdmission + "." + terraformConfig.ResourcePrefix + "]")},
		}
	}

	clusterBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnTemp)
	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(defaults.DefaultPodSecurityAdmission, cty.StringVal(psact))

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		v2.SetProxyConfig(clusterBlockBody, terraformConfig)
	}

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	rkeConfigBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(k8sVersion))

	networkBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Network, nil)
	networkBlockBody := networkBlock.Body()

	networkBlockBody.SetAttributeValue(defaults.Plugin, cty.StringVal(terraformConfig.CNI))

	rootBody.AppendNewline()

	if terraformConfig.PrivateRegistries != nil && strings.Contains(terraformConfig.Module, modules.EC2) {
		registryBlock := rkeConfigBlockBody.AppendNewBlock(defaults.RKE1PrivateRegistries, nil)
		registryBlockBody := registryBlock.Body()

		setRKE1PrivateRegistryConfig(registryBlockBody, terraformConfig)

		rootBody.AppendNewline()
	}

	if terraformConfig.ETCDRKE1 != nil {
		servicesBlock := rkeConfigBlockBody.AppendNewBlock(defaults.Services, nil)
		servicesBlockBody := servicesBlock.Body()

		setEtcdConfig(servicesBlockBody, terraformConfig)

		rootBody.AppendNewline()
	}

	clusterSyncNodePoolIDs := ""

	for count, pool := range nodePools {
		setNodePool(nodePools, count, pool, rootBody, clusterSyncNodePoolIDs, terraformConfig)
	}

	setClusterSync(rootBody, clusterSyncNodePoolIDs, terraformConfig.ResourcePrefix)

	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE1 configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}
