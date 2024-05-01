package rke2k3s

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
)

// SetRKE2K3s is a function that will set the RKE2/K3S configurations in the main.tf file.
func SetRKE2K3s(clusterName, poolName, k8sVersion, psact string, nodePools []config.Nodepool, snapshots config.Snapshots, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	newFile, rootBody := resources.SetProvidersAndUsersTF(rancherConfig, terraformConfig)

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
		setMachinePool(nodePools, count, pool, rkeConfigBlockBody, poolName)
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

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
