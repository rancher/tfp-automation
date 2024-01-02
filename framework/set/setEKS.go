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

// SetEKS is a function that will set the EKS configurations in the main.tf file.
func SetEKS(clusterName, k8sVersion string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig("rancher", rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig("terraform", terraformConfig)

	newFile, rootBody := setProvidersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	ec2CredConfigBlock := cloudCredBlockBody.AppendNewBlock("amazonec2_credential_config", nil)
	ec2CredConfigBlockBody := ec2CredConfigBlock.Body()

	ec2CredConfigBlockBody.SetAttributeValue("access_key", cty.StringVal(terraformConfig.AWSAccessKey))
	ec2CredConfigBlockBody.SetAttributeValue("secret_key", cty.StringVal(terraformConfig.AWSSecretKey))
	ec2CredConfigBlockBody.SetAttributeValue("default_region", cty.StringVal(terraformConfig.Region))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cluster", "rancher2_cluster"})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue("name", cty.StringVal(clusterName))

	eksConfigBlock := clusterBlockBody.AppendNewBlock("eks_config_v2", nil)
	eksConfigBlockBody := eksConfigBlock.Body()

	cloudCredID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`rancher2_cloud_credential.rancher2_cloud_credential.id`)},
	}

	eksConfigBlockBody.SetAttributeRaw("cloud_credential_id", cloudCredID)
	eksConfigBlockBody.SetAttributeValue("region", cty.StringVal(terraformConfig.Region))
	eksConfigBlockBody.SetAttributeValue("kubernetes_version", cty.StringVal(k8sVersion))
	awsSubnetsList := format.ListOfStrings(terraformConfig.AWSSubnets)
	eksConfigBlockBody.SetAttributeRaw("subnets", awsSubnetsList)
	awsSecGroupsList := format.ListOfStrings(terraformConfig.AWSSecurityGroups)
	eksConfigBlockBody.SetAttributeRaw("security_groups", awsSecGroupsList)
	eksConfigBlockBody.SetAttributeValue("private_access", cty.BoolVal(terraformConfig.PrivateAccess))
	eksConfigBlockBody.SetAttributeValue("public_access", cty.BoolVal(terraformConfig.PublicAccess))

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		SetResourceNodepoolValidation(pool, poolNum)


		nodePoolsBlock := eksConfigBlockBody.AppendNewBlock("node_groups", nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.HostnamePrefix+`-pool`+poolNum))
		nodePoolsBlockBody.SetAttributeValue("instance_type", cty.StringVal(pool.InstanceType))
		nodePoolsBlockBody.SetAttributeValue("desired_size", cty.NumberIntVal(pool.DesiredSize))
		nodePoolsBlockBody.SetAttributeValue("max_size", cty.NumberIntVal(pool.MaxSize))
		nodePoolsBlockBody.SetAttributeValue("min_size", cty.NumberIntVal(pool.MinSize))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write EKS configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
