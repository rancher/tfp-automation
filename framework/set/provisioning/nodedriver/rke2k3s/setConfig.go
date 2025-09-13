package rke2k3s

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	aws "github.com/rancher/tfp-automation/framework/set/provisioning/providers/aws"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	harvester "github.com/rancher/tfp-automation/framework/set/provisioning/providers/harvester"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
)

const (
	clusterV2       = "rancher2_cluster_v2"
	machineConfigV2 = "rancher2_machine_config_v2"

	cloudCredentialName       = "cloud_credential_name"
	cloudCredentialSecretName = "cloud_credential_secret_name"
	controlPlaneRole          = "control_plane_role"
	etcdRole                  = "etcd_role"
	workerRole                = "worker_role"

	upgradeStrategy         = "upgrade_strategy"
	controlPlaneConcurrency = "control_plane_concurrency"
	workerConcurrency       = "worker_concurrency"

	disableSnapshots     = "disable_snapshots"
	snapshotScheduleCron = "snapshot_schedule_cron"
	snapshotRetention    = "snapshot_retention"
	s3Config             = "s3_config"
	bucket               = "bucket"
	endpointCA           = "endpoint_ca"
	skipSSLVerify        = "skip_ssl_verify"

	hostname              = "hostname"
	authConfigSecretName  = "auth_config_secret_name"
	tlsSecretName         = "tls_secret_name"
	caBundleName          = "ca_bundle"
	insecure              = "insecure"
	systemDefaultRegistry = "system-default-registry"
	project               = "project"
	endpoints             = "endpoints"
	rewrites              = "rewrites"

	httpProxy    = "HTTP_PROXY"
	httpsProxy   = "HTTPS_PROXY"
	noProxy      = "NO_PROXY"
	noProxyValue = "localhost,127.0.0.0/8,10.0.0.0/8,172.0.0.0/8,192.168.0.0/16,.svc,.cluster.local,cattle-system.svc,169.254.169.254"
)

// SetRKE2K3s is a function that will set the RKE2/K3S configurations in the main.tf file.
func SetRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File, rbacRole config.Role) (*hclwrite.File, *os.File, error) {
	switch terraformConfig.Module {
	case modules.EC2RKE2, modules.EC2K3s:
		aws.SetAWSRKE2K3SProvider(rootBody, terraformConfig)
	case modules.AzureRKE2, modules.AzureK3s:
		azure.SetAzureRKE2K3SProvider(rootBody, terraformConfig)
	case modules.HarvesterRKE2, modules.HarvesterK3s:
		harvester.SetHarvesterCredentialProvider(rootBody, terraformConfig)
	case modules.LinodeRKE2, modules.LinodeK3s:
		linode.SetLinodeRKE2K3SProvider(rootBody, terraformConfig)
	case modules.VsphereRKE2, modules.VsphereK3s:
		vsphere.SetVsphereRKE2K3SProvider(rootBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(terratestConfig.PSACT, defaults.RancherBaseline) {
		rootBody, err := resources.SetBaselinePSACT(newFile, rootBody, terraformConfig.ResourcePrefix)
		if err != nil {
			return nil, nil, err
		}

		rootBody.AppendNewline()
	}

	machineConfigBlockBody, err := setMachineConfig(rootBody, terraformConfig, terratestConfig.PSACT)
	if err != nil {
		return nil, nil, err
	}

	switch terraformConfig.Module {
	case modules.EC2RKE2, modules.EC2K3s:
		aws.SetAWSRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case modules.AzureRKE2, modules.AzureK3s:
		azure.SetAzureRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case modules.HarvesterRKE2, modules.HarvesterK3s:
		harvester.SetHarvesterRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case modules.LinodeRKE2, modules.LinodeK3s:
		linode.SetLinodeRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case modules.VsphereRKE2, modules.VsphereK3s:
		vsphere.SetVsphereRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	clusterBlockBody, err := setClusterConfig(rootBody, terraformConfig, terratestConfig.PSACT, terratestConfig.KubernetesVersion)
	if err != nil {
		return nil, nil, err
	}

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		err = SetProxyConfig(clusterBlockBody, terraformConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	rkeConfigBlockBody, err := setRKEConfig(clusterBlockBody, terraformConfig)
	if err != nil {
		return nil, nil, err
	}

	for count, pool := range terratestConfig.Nodepools {
		err = setMachinePool(terraformConfig, count, pool, rkeConfigBlockBody)
		if err != nil {
			return nil, nil, err
		}
	}

	if terraformConfig.PrivateRegistries != nil && strings.Contains(terraformConfig.Module, modules.EC2) {
		if terraformConfig.PrivateRegistries.Username != "" {
			rootBody.AppendNewline()
			CreateRegistrySecret(terraformConfig, rootBody)
		}

		if terraformConfig.PrivateRegistries.SystemDefaultRegistry != "" {
			err = SetMachineSelectorConfig(rkeConfigBlockBody, terraformConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		err = SetPrivateRegistryConfig(rkeConfigBlockBody, terraformConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	if terraformConfig.ETCD != nil {
		err = setEtcdConfig(rkeConfigBlockBody, terraformConfig)
		if err != nil {
			return nil, nil, err
		}
	}

	if terratestConfig.SnapshotInput.CreateSnapshot {
		err = SetCreateRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody)
		if err != nil {
			return nil, nil, err
		}
	}

	if terratestConfig.SnapshotInput.RestoreSnapshot {
		err = SetRestoreRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody, terratestConfig.SnapshotInput)
		if err != nil {
			return nil, nil, err
		}
	}

	rootBody.AppendNewline()

	return newFile, file, nil
}
