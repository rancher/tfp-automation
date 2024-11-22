package rke2k3s

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	aws "github.com/rancher/tfp-automation/framework/set/provisioning/providers/aws"
	azure "github.com/rancher/tfp-automation/framework/set/provisioning/providers/azure"
	linode "github.com/rancher/tfp-automation/framework/set/provisioning/providers/linode"
	vsphere "github.com/rancher/tfp-automation/framework/set/provisioning/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/rbac"
	"github.com/rancher/tfp-automation/framework/set/resources"
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
)

// SetRKE2K3s is a function that will set the RKE2/K3S configurations in the main.tf file.
func SetRKE2K3s(client *rancher.Client, terraformConfig *config.TerraformConfig, clusterName, poolName, k8sVersion, psact string,
	nodePools []config.Nodepool, snapshots config.Snapshots, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, rbacRole config.Role) error {
	switch {
	case terraformConfig.Module == modules.EC2RKE2 || terraformConfig.Module == modules.EC2K3s:
		aws.SetAWSRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == modules.AzureRKE2 || terraformConfig.Module == modules.AzureK3s:
		azure.SetAzureRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == modules.LinodeRKE2 || terraformConfig.Module == modules.LinodeK3s:
		linode.SetLinodeRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == modules.VsphereRKE2 || terraformConfig.Module == modules.VsphereK3s:
		vsphere.SetVsphereRKE2K3SProvider(rootBody, terraformConfig)
	}

	rootBody.AppendNewline()

	if strings.Contains(psact, defaults.RancherBaseline) {
		newFile, rootBody = resources.SetBaselinePSACT(newFile, rootBody)

		rootBody.AppendNewline()
	}

	machineConfigBlock := rootBody.AppendNewBlock(defaults.Resource, []string{machineConfigV2, machineConfigV2})
	machineConfigBlockBody := machineConfigBlock.Body()

	if psact == defaults.RancherBaseline {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + defaults.PodSecurityAdmission + "." +
				defaults.PodSecurityAdmission + "]")},
		}

		machineConfigBlockBody.SetAttributeRaw(defaults.DependsOn, dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue(defaults.GenerateName, cty.StringVal(terraformConfig.MachineConfigName))

	switch {
	case terraformConfig.Module == modules.EC2RKE2 || terraformConfig.Module == modules.EC2K3s:
		aws.SetAWSRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == modules.AzureRKE2 || terraformConfig.Module == modules.AzureK3s:
		azure.SetAzureRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == modules.LinodeRKE2 || terraformConfig.Module == modules.LinodeK3s:
		linode.SetLinodeRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == modules.VsphereRKE2 || terraformConfig.Module == modules.VsphereK3s:
		vsphere.SetVsphereRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{clusterV2, clusterV2})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(k8sVersion))
	clusterBlockBody.SetAttributeValue(defaults.EnableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue(defaults.DefaultPodSecurityAdmission, cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue(defaults.DefaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(defaults.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	for count, pool := range nodePools {
		setMachinePool(terraformConfig, count, pool, rkeConfigBlockBody, poolName)
	}

	if terraformConfig.PrivateRegistries != nil && strings.Contains(terraformConfig.Module, modules.EC2) {
		setMachineSelectorConfig(rkeConfigBlockBody, terraformConfig)

		registryBlock := rkeConfigBlockBody.AppendNewBlock(defaults.PrivateRegistries, nil)
		registryBlockBody := registryBlock.Body()

		setPrivateRegistryConfig(registryBlockBody, terraformConfig)
	}

	upgradeStrategyBlock := rkeConfigBlockBody.AppendNewBlock(upgradeStrategy, nil)
	upgradeStrategyBlockBody := upgradeStrategyBlock.Body()

	upgradeStrategyBlockBody.SetAttributeValue(controlPlaneConcurrency, cty.StringVal(("10%")))
	upgradeStrategyBlockBody.SetAttributeValue(workerConcurrency, cty.StringVal(("10%")))

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

	if rbacRole != "" {
		user, err := rbac.SetUsers(newFile, rootBody, rbacRole)
		if err != nil {
			return err
		}

		rootBody.AppendNewline()

		if strings.Contains(string(rbacRole), project) {
			rbac.AddProjectMember(client, clusterName, newFile, rootBody, nil, rbacRole, user, false)
		} else {
			rbac.AddClusterRole(client, clusterName, newFile, rootBody, nil, rbacRole, user, false)
		}
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
