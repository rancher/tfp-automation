package framework

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetRKE2K3s is a function that will set the RKE2/K3S configurations in the main.tf file.
func SetRKE2K3s(clusterName, k8sVersion, psact string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig("rancher", rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig("terraform", terraformConfig)

	newFile, rootBody := setProvidersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	if terraformConfig.Module == ec2RKE2 || terraformConfig.Module == ec2K3s {
		cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
		cloudCredBlockBody := cloudCredBlock.Body()

		cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

		ec2CredBlock := cloudCredBlockBody.AppendNewBlock("amazonec2_credential_config", nil)
		ec2CredBlockBody := ec2CredBlock.Body()

		ec2CredBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.AWSAccessKey))
		ec2CredBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.AWSSecretKey))

		rootBody.AppendNewline()
	}

	if terraformConfig.Module == linodeRKE2 || terraformConfig.Module == linodeK3s {
		cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
		cloudCredBlockBody := cloudCredBlock.Body()

		cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

		linodeCredBlock := cloudCredBlockBody.AppendNewBlock("linode_credential_config", nil)
		linodeCredBlockBody := linodeCredBlock.Body()

		linodeCredBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeToken))

		rootBody.AppendNewline()
	}

	machineConfigBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_machine_config_v2", "rancher2_machine_config_v2"})
	machineConfigBlockBody := machineConfigBlock.Body()

	machineConfigBlockBody.SetAttributeValue("generate_name", cty.StringVal(terraformConfig.MachineConfigName))

	if terraformConfig.Module == ec2RKE2 || terraformConfig.Module == ec2K3s {
		ec2ConfigBlock := machineConfigBlockBody.AppendNewBlock("amazonec2_config", nil)
		ec2ConfigBlockBody := ec2ConfigBlock.Body()

		ec2ConfigBlockBody.SetAttributeValue("ami", cty.StringVal(terraformConfig.Ami))
		ec2ConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.Region))
		awsSecGroupNames := format.ListOfStrings(terraformConfig.AWSSecurityGroupNames)
		ec2ConfigBlockBody.SetAttributeRaw("security_group", awsSecGroupNames)
		ec2ConfigBlockBody.SetAttributeValue("subnet_id", cty.StringVal(terraformConfig.AWSSubnetID))
		ec2ConfigBlockBody.SetAttributeValue("vpc_id", cty.StringVal(terraformConfig.AWSVpcID))
		ec2ConfigBlockBody.SetAttributeValue("zone", cty.StringVal(terraformConfig.AWSZoneLetter))
	}

	if terraformConfig.Module == linodeRKE2 || terraformConfig.Module == linodeK3s {
		linodeConfigBlock := machineConfigBlockBody.AppendNewBlock("linode_config", nil)
		linodeConfigBlockBody := linodeConfigBlock.Body()

		linodeConfigBlockBody.SetAttributeValue("image", cty.StringVal(terraformConfig.LinodeImage))
		linodeConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.Region))
		linodeConfigBlockBody.SetAttributeValue("root_pass", cty.StringVal(terraformConfig.LinodeRootPass))
		linodeConfigBlockBody.SetAttributeValue("token", cty.StringVal(terraformConfig.LinodeToken))
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

		SetResourceNodepoolValidation(pool, poolNum)

		machinePoolsBlock := rkeConfigBlockBody.AppendNewBlock("machine_pools", nil)
		machinePoolsBlockBody := machinePoolsBlock.Body()

		machinePoolsBlockBody.SetAttributeValue("name", cty.StringVal(`pool`+poolNum))

		cloudCredSecretName := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`rancher2_cloud_credential.rancher2_cloud_credential.id`)},
		}

		machinePoolsBlockBody.SetAttributeRaw("cloud_credential_secret_name", cloudCredSecretName)
		machinePoolsBlockBody.SetAttributeValue("control_plane_role", cty.BoolVal(pool.Controlplane))
		machinePoolsBlockBody.SetAttributeValue("etcd_role", cty.BoolVal(pool.Etcd))
		machinePoolsBlockBody.SetAttributeValue("worker_role", cty.BoolVal(pool.Worker))
		machinePoolsBlockBody.SetAttributeValue("quantity", cty.NumberIntVal(pool.Quantity))

		machineConfigBlock := machinePoolsBlockBody.AppendNewBlock("machine_config", nil)
		machineConfigBlockBody := machineConfigBlock.Body()

		kind := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`rancher2_machine_config_v2.rancher2_machine_config_v2.kind`)},
		}

		machineConfigBlockBody.SetAttributeRaw("kind", kind)

		name := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(`rancher2_machine_config_v2.rancher2_machine_config_v2.name`)},
		}

		machineConfigBlockBody.SetAttributeRaw("name", name)

		count++
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write RKE2/K3S configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
