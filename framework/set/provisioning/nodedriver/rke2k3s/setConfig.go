package rke2k3s

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
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
func SetRKE2K3s(terraformConfig *config.TerraformConfig, k8sVersion, psact string, nodePools []config.Nodepool,
	snapshots config.Snapshots, newFile *hclwrite.File, rootBody *hclwrite.Body,
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

	if strings.Contains(psact, defaults.RancherBaseline) {
		newFile, rootBody = resources.SetBaselinePSACT(newFile, rootBody, terraformConfig.ResourcePrefix)

		rootBody.AppendNewline()
	}

	machineConfigBlock := rootBody.AppendNewBlock(defaults.Resource, []string{machineConfigV2, terraformConfig.ResourcePrefix})
	machineConfigBlockBody := machineConfigBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Rancher2 + "." + defaults.StandardUser)},
	}

	machineConfigBlockBody.SetAttributeRaw(defaults.Provider, provider)

	if psact == defaults.RancherBaseline {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + defaults.PodSecurityAdmission + "." +
				terraformConfig.ResourcePrefix + "]")},
		}

		machineConfigBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue(defaults.GenerateName, cty.StringVal(terraformConfig.ResourcePrefix))

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

	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{clusterV2, terraformConfig.ResourcePrefix})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeRaw(defaults.Provider, provider)
	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))
	clusterBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(k8sVersion))
	clusterBlockBody.SetAttributeValue(defaults.EnableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue(defaults.DefaultPodSecurityAdmission, cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue(defaults.DefaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	if terraformConfig.Proxy != nil && terraformConfig.Proxy.ProxyBastion != "" {
		SetProxyConfig(clusterBlockBody, terraformConfig)
	}

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	if terraformConfig.ChartValues != "" {
		chartValues := hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOF\n" + terraformConfig.ChartValues + "\nEOF"},
		})

		rkeConfigBlockBody.SetAttributeRaw(defaults.ChartValues, chartValues)
	}

	machineGlobalConfigValue := hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\ncni: " + terraformConfig.CNI + "\ndisable-kube-proxy: " + terraformConfig.DisableKubeProxy + "\nEOF"},
	})

	rkeConfigBlockBody.SetAttributeRaw(defaults.MachineGlobalConfig, machineGlobalConfigValue)

	for count, pool := range nodePools {
		setMachinePool(terraformConfig, count, pool, rkeConfigBlockBody)
	}

	if terraformConfig.PrivateRegistries != nil && strings.Contains(terraformConfig.Module, modules.EC2) {
		if terraformConfig.PrivateRegistries.Username != "" {
			rootBody.AppendNewline()
			CreateRegistrySecret(terraformConfig, rootBody)
		}

		if terraformConfig.PrivateRegistries.SystemDefaultRegistry != "" {
			SetMachineSelectorConfig(rkeConfigBlockBody, terraformConfig)
		}

		registryBlock := rkeConfigBlockBody.AppendNewBlock(defaults.PrivateRegistries, nil)
		registryBlockBody := registryBlock.Body()

		SetPrivateRegistryConfig(registryBlockBody, terraformConfig)
	}

	if terraformConfig.ETCD != nil {
		setEtcdConfig(rkeConfigBlockBody, terraformConfig)
	}

	if snapshots.CreateSnapshot {
		SetCreateRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody)
	}

	if snapshots.RestoreSnapshot {
		SetRestoreRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody, snapshots)
	}

	rootBody.AppendNewline()

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3s configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}
