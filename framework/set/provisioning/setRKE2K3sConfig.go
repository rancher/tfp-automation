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

	newFile, rootBody := SetProvidersAndUsersTF(rancherConfig, terraformConfig)

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

	if strings.Contains(psact, rancherBaseline) {
		newFile, rootBody = SetBaselinePSACT(newFile, rootBody)

		rootBody.AppendNewline()
	}

	machineConfigBlock := rootBody.AppendNewBlock(resource, []string{machineConfigV2, machineConfigV2})
	machineConfigBlockBody := machineConfigBlock.Body()

	if psact == rancherBaseline {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + podSecurityAdmission + "." +
				podSecurityAdmission + "]")},
		}

		machineConfigBlockBody.SetAttributeRaw(dependsOn, dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue(generateName, cty.StringVal(terraformConfig.MachineConfigName))

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

	clusterBlock := rootBody.AppendNewBlock(resource, []string{clusterV2, clusterV2})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(resourceName, cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue(kubernetesVersion, cty.StringVal(k8sVersion))
	clusterBlockBody.SetAttributeValue(enableNetworkPolicy, cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue(defaultPodSecurityAdmission, cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue(defaultClusterRoleForProjectMembers, cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock(rkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		machinePoolsBlock := rkeConfigBlockBody.AppendNewBlock(machinePools, nil)
		machinePoolsBlockBody := machinePoolsBlock.Body()

		machinePoolsBlockBody.SetAttributeValue(resourceName, cty.StringVal(poolName+poolNum))

		cloudCredSecretName := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(cloudCredential + "." + cloudCredential + ".id")},
		}

		machinePoolsBlockBody.SetAttributeRaw(cloudCredentialSecretName, cloudCredSecretName)
		machinePoolsBlockBody.SetAttributeValue(controlPlaneRole, cty.BoolVal(pool.Controlplane))
		machinePoolsBlockBody.SetAttributeValue(etcdRole, cty.BoolVal(pool.Etcd))
		machinePoolsBlockBody.SetAttributeValue(workerRole, cty.BoolVal(pool.Worker))
		machinePoolsBlockBody.SetAttributeValue(quantity, cty.NumberIntVal(pool.Quantity))

		machineConfigBlock := machinePoolsBlockBody.AppendNewBlock(machineConfig, nil)
		machineConfigBlockBody := machineConfigBlock.Body()

		kind := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(machineConfigV2 + "." + machineConfigV2 + ".kind")},
		}

		machineConfigBlockBody.SetAttributeRaw(resourceKind, kind)

		name := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(machineConfigV2 + "." + machineConfigV2 + ".name")},
		}

		machineConfigBlockBody.SetAttributeRaw(resourceName, name)
	}

	upgradeStrategyBlock := rkeConfigBlockBody.AppendNewBlock(upgradeStrategy, nil)
	upgradeStrategyBlockBody := upgradeStrategyBlock.Body()

	upgradeStrategyBlockBody.SetAttributeValue(controlPlaneConcurrency, cty.StringVal(("10%")))
	upgradeStrategyBlockBody.SetAttributeValue(workerConcurrency, cty.StringVal(("10%")))

	if terraformConfig.ETCD != nil {
		snapshotBlock := rkeConfigBlockBody.AppendNewBlock(etcd, nil)
		snapshotBlockBody := snapshotBlock.Body()

		snapshotBlockBody.SetAttributeValue(disableSnapshots, cty.BoolVal(terraformConfig.ETCD.DisableSnapshots))
		snapshotBlockBody.SetAttributeValue(snapshotScheduleCron, cty.StringVal(terraformConfig.ETCD.SnapshotScheduleCron))
		snapshotBlockBody.SetAttributeValue(snapshotRetention, cty.NumberIntVal(int64(terraformConfig.ETCD.SnapshotRetention)))

		if strings.Contains(terraformConfig.Module, modules.EC2) && terraformConfig.ETCD.S3 != nil {
			s3ConfigBlock := snapshotBlockBody.AppendNewBlock(s3Config, nil)
			s3ConfigBlockBody := s3ConfigBlock.Body()

			cloudCredSecretName := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte(cloudCredential + "." + cloudCredential + ".id")},
			}

			s3ConfigBlockBody.SetAttributeValue(bucket, cty.StringVal(terraformConfig.ETCD.S3.Bucket))
			s3ConfigBlockBody.SetAttributeValue(endpoint, cty.StringVal(terraformConfig.ETCD.S3.Endpoint))
			s3ConfigBlockBody.SetAttributeRaw(cloudCredentialName, cloudCredSecretName)
			s3ConfigBlockBody.SetAttributeValue(endpointCA, cty.StringVal(terraformConfig.ETCD.S3.EndpointCA))
			s3ConfigBlockBody.SetAttributeValue(folder, cty.StringVal(terraformConfig.ETCD.S3.Folder))
			s3ConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.ETCD.S3.Region))
			s3ConfigBlockBody.SetAttributeValue(skipSSLVerify, cty.BoolVal(terraformConfig.ETCD.S3.SkipSSLVerify))
		}
	}

	if snapshots.CreateSnapshot {
		setCreateRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody)
	}

	if snapshots.RestoreSnapshot {
		setRestoreRKE2K3SSnapshot(terraformConfig, rkeConfigBlockBody, snapshots)
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
