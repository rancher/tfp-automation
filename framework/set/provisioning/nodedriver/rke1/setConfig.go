package rke1

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	v2 "github.com/rancher/tfp-automation/framework/set/provisioning/nodedriver/rke2k3s"
	aws "github.com/rancher/tfp-automation/framework/set/provisioning/providers/aws"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	harvester "github.com/rancher/tfp-automation/framework/set/provisioning/providers/harvester"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
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
func SetRKE1(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File, rbacRole config.Role) (*hclwrite.File, *os.File, error) {
	nodeTemplateBlock := rootBody.AppendNewBlock(general.Resource, []string{nodeTemplate, terraformConfig.ResourcePrefix})
	nodeTemplateBlockBody := nodeTemplateBlock.Body()

	nodeTemplateBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	if terraformConfig.PrivateRegistries != nil {
		nodeTemplateBlockBody.SetAttributeValue(clusters.EngineInsecureRegistry, cty.ListVal([]cty.Value{
			cty.StringVal(terraformConfig.PrivateRegistries.URL),
		}))
	}

	switch terraformConfig.Module {
	case modules.EC2RKE1:
		aws.SetAWSRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case modules.AzureRKE1:
		azure.SetAzureRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case modules.LinodeRKE1:
		linode.SetLinodeRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case modules.HarvesterRKE1:
		harvester.SetHarvesterCredentialProvider(rootBody, terraformConfig)
		harvester.SetHarvesterRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	case modules.VsphereRKE1:
		vsphere.SetVsphereRKE1Provider(nodeTemplateBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(terratestConfig.PSACT, clusters.RancherBaseline) {
		rootBody, err := resources.SetBaselinePSACT(newFile, rootBody, terraformConfig.ResourcePrefix)
		if err != nil {
			return nil, nil, err
		}

		rootBody.AppendNewline()
	}

	clusterBlockBody, err := setClusterConfig(rootBody, terraformConfig, terratestConfig.PSACT)
	if err != nil {
		return nil, nil, err
	}

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		err = v2.SetProxyConfig(clusterBlockBody, terraformConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	rkeConfigBlockBody, err := setRKEConfig(clusterBlockBody, terraformConfig, terratestConfig.KubernetesVersion)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	if terraformConfig.PrivateRegistries != nil && strings.Contains(terraformConfig.Module, modules.EC2) {
		err = setRKE1PrivateRegistryConfig(rkeConfigBlockBody, terraformConfig)
		if err != nil {
			return nil, nil, err
		}

		rootBody.AppendNewline()
	}

	if terraformConfig.ETCDRKE1 != nil {
		err = setEtcdConfig(rkeConfigBlockBody, terraformConfig)
		if err != nil {
			return nil, nil, err
		}

		rootBody.AppendNewline()
	}

	clusterSyncNodePoolIDs := ""

	for count, pool := range terratestConfig.Nodepools {
		err = setNodePool(terratestConfig.Nodepools, count, pool, rootBody, clusterSyncNodePoolIDs, terraformConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	err = setClusterSync(rootBody, clusterSyncNodePoolIDs, terraformConfig.ResourcePrefix)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	return newFile, file, nil
}
