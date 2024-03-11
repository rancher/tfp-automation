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

// SetRKE2K3s is a function that will set the RKE2/K3S configurations in the main.tf file.
func SetRKE2K3s(clusterName, k8sVersion, psact string, nodePools []config.Nodepool, snapshots config.Snapshots, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig("rancher", rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(config.TerraformConfigurationFileKey, terraformConfig)

	terratestConfig := new(config.TerratestConfig)
	ranchFrame.LoadConfig(config.TerratestConfigurationFileKey, terratestConfig)

	newFile, rootBody := SetProvidersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	switch {
	case terraformConfig.Module == azure_rke2 || terraformConfig.Module == azure_k3s:
		azure.SetAzureRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == ec2RKE2 || terraformConfig.Module == ec2K3s:
		ec2.SetEC2RKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == linodeRKE2 || terraformConfig.Module == linodeK3s:
		linode.SetLinodeRKE2K3SProvider(rootBody, terraformConfig)
	case terraformConfig.Module == vsphereRKE2 || terraformConfig.Module == vsphereK3s:
		vsphere.SetVsphereRKE2K3SProvider(rootBody, terraformConfig)
	}
	rootBody.AppendNewline()

	if strings.Contains(psact, rancherBaseline) {
		newFile, rootBody = SetBaselinePSACT(newFile, rootBody)
		
		rootBody.AppendNewline()
	}

	machineConfigBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_machine_config_v2", "rancher2_machine_config_v2"})
	machineConfigBlockBody := machineConfigBlock.Body()
	
	if psact == "rancher-baseline" {
		dependsOnTemp := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("[rancher2_pod_security_admission_configuration_template." + 
			"rancher2_pod_security_admission_configuration_template]")},
		}
	
		machineConfigBlockBody.SetAttributeRaw("depends_on", dependsOnTemp)
	}

	machineConfigBlockBody.SetAttributeValue("generate_name", cty.StringVal(terraformConfig.MachineConfigName))

	switch {
	case terraformConfig.Module == azure_rke2 || terraformConfig.Module == azure_k3s:
		azure.SetAzureRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == ec2RKE2 || terraformConfig.Module == ec2K3s:
		ec2.SetEC2RKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == linodeRKE2 || terraformConfig.Module == linodeK3s:
		linode.SetLinodeRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	case terraformConfig.Module == vsphereRKE2 || terraformConfig.Module == vsphereK3s:
		vsphere.SetVsphereRKE2K3SMachineConfig(machineConfigBlockBody, terraformConfig)
	}

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cluster_v2", "rancher2_cluster_v2"})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue("name", cty.StringVal(clusterName))
	clusterBlockBody.SetAttributeValue("kubernetes_version", cty.StringVal(k8sVersion))
	clusterBlockBody.SetAttributeValue("enable_network_policy", cty.BoolVal(terraformConfig.EnableNetworkPolicy))
	clusterBlockBody.SetAttributeValue("default_pod_security_admission_configuration_template_name", cty.StringVal(psact))
	clusterBlockBody.SetAttributeValue("default_cluster_role_for_project_members", cty.StringVal(terraformConfig.DefaultClusterRoleForProjectMembers))

	rkeConfigBlock := clusterBlockBody.AppendNewBlock("rke_config", nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		machinePoolsBlock := rkeConfigBlockBody.AppendNewBlock("machine_pools", nil)
		machinePoolsBlockBody := machinePoolsBlock.Body()

		machinePoolsBlockBody.SetAttributeValue("name", cty.StringVal("tfp-pool"+poolNum))

		cloudCredSecretName := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_cloud_credential.rancher2_cloud_credential.id")},
		}

		machinePoolsBlockBody.SetAttributeRaw("cloud_credential_secret_name", cloudCredSecretName)
		machinePoolsBlockBody.SetAttributeValue("control_plane_role", cty.BoolVal(pool.Controlplane))
		machinePoolsBlockBody.SetAttributeValue("etcd_role", cty.BoolVal(pool.Etcd))
		machinePoolsBlockBody.SetAttributeValue("worker_role", cty.BoolVal(pool.Worker))
		machinePoolsBlockBody.SetAttributeValue("quantity", cty.NumberIntVal(pool.Quantity))

		machineConfigBlock := machinePoolsBlockBody.AppendNewBlock("machine_config", nil)
		machineConfigBlockBody := machineConfigBlock.Body()

		kind := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_machine_config_v2.rancher2_machine_config_v2.kind")},
		}

		machineConfigBlockBody.SetAttributeRaw("kind", kind)

		name := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_machine_config_v2.rancher2_machine_config_v2.name")},
		}

		machineConfigBlockBody.SetAttributeRaw("name", name)

		count++
	}

	upgradeStrategyBlock := rkeConfigBlockBody.AppendNewBlock("upgrade_strategy", nil)
	upgradeStrategyBlockBody := upgradeStrategyBlock.Body()

	upgradeStrategyBlockBody.SetAttributeValue("control_plane_concurrency", cty.StringVal(("10%")))
	upgradeStrategyBlockBody.SetAttributeValue("worker_concurrency", cty.StringVal(("10%")))

	if terraformConfig.ETCD != nil {
		snapshotBlock := rkeConfigBlockBody.AppendNewBlock("etcd", nil)
		snapshotBlockBody := snapshotBlock.Body()

		snapshotBlockBody.SetAttributeValue("disable_snapshots", cty.BoolVal(terraformConfig.ETCD.DisableSnapshots))
		snapshotBlockBody.SetAttributeValue("snapshot_schedule_cron", cty.StringVal(terraformConfig.ETCD.SnapshotScheduleCron))
		snapshotBlockBody.SetAttributeValue("snapshot_retention", cty.NumberIntVal(int64(terraformConfig.ETCD.SnapshotRetention)))

		if strings.Contains(terraformConfig.Module, "ec2") && terraformConfig.ETCD.S3 != nil {
			s3ConfigBlock := snapshotBlockBody.AppendNewBlock("s3_config", nil)
			s3ConfigBlockBody := s3ConfigBlock.Body()

			cloudCredSecretName := hclwrite.Tokens{
				{Type: hclsyntax.TokenIdent, Bytes: []byte("rancher2_cloud_credential.rancher2_cloud_credential.id")},
			}

			s3ConfigBlockBody.SetAttributeValue("bucket", cty.StringVal(terraformConfig.ETCD.S3.Bucket))
			s3ConfigBlockBody.SetAttributeValue("endpoint", cty.StringVal(terraformConfig.ETCD.S3.Endpoint))
			s3ConfigBlockBody.SetAttributeRaw("cloud_credential_name", cloudCredSecretName)
			s3ConfigBlockBody.SetAttributeValue("endpoint_ca", cty.StringVal(terraformConfig.ETCD.S3.EndpointCA))
			s3ConfigBlockBody.SetAttributeValue("folder", cty.StringVal(terraformConfig.ETCD.S3.Folder))
			s3ConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.ETCD.S3.Region))
			s3ConfigBlockBody.SetAttributeValue("skip_ssl_verify", cty.BoolVal(terraformConfig.ETCD.S3.SkipSSLVerify))
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
