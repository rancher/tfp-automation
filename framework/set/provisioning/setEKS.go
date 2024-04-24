package provisioning

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	ranchFrame "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/amazon"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetEKS is a function that will set the EKS configurations in the main.tf file.
func SetEKS(clusterName, k8sVersion string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	ranchFrame.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	ranchFrame.LoadConfig(configs.Terraform, terraformConfig)

	newFile, rootBody := SetProvidersAndUsersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock(resource, []string{cloudCredential, cloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.CloudCredentialName))

	ec2CredConfigBlock := cloudCredBlockBody.AppendNewBlock(amazon.EC2CredentialConfig, nil)
	ec2CredConfigBlockBody := ec2CredConfigBlock.Body()

	ec2CredConfigBlockBody.SetAttributeValue(accessKey, cty.StringVal(terraformConfig.AWSConfig.AWSAccessKey))
	ec2CredConfigBlockBody.SetAttributeValue(secretKey, cty.StringVal(terraformConfig.AWSConfig.AWSSecretKey))
	ec2CredConfigBlockBody.SetAttributeValue(amazon.DefaultRegion, cty.StringVal(terraformConfig.AWSConfig.Region))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(resource, []string{cluster, cluster})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(resourceName, cty.StringVal(clusterName))

	eksConfigBlock := clusterBlockBody.AppendNewBlock(amazon.EKSConfig, nil)
	eksConfigBlockBody := eksConfigBlock.Body()

	cloudCredID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(cloudCredential + "." + cloudCredential + ".id")},
	}

	eksConfigBlockBody.SetAttributeRaw(cloudCredentialID, cloudCredID)
	eksConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.AWSConfig.Region))
	eksConfigBlockBody.SetAttributeValue(kubernetesVersion, cty.StringVal(k8sVersion))
	awsSubnetsList := format.ListOfStrings(terraformConfig.AWSConfig.AWSSubnets)
	eksConfigBlockBody.SetAttributeRaw(amazon.Subnets, awsSubnetsList)
	awsSecGroupsList := format.ListOfStrings(terraformConfig.AWSConfig.AWSSecurityGroups)
	eksConfigBlockBody.SetAttributeRaw(amazon.SecurityGroups, awsSecGroupsList)
	eksConfigBlockBody.SetAttributeValue(amazon.PrivateAccess, cty.BoolVal(terraformConfig.AWSConfig.PrivateAccess))
	eksConfigBlockBody.SetAttributeValue(amazon.PublicAccess, cty.BoolVal(terraformConfig.AWSConfig.PublicAccess))

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolsBlock := eksConfigBlockBody.AppendNewBlock(amazon.NodeGroups, nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.HostnamePrefix+`-pool`+poolNum))
		nodePoolsBlockBody.SetAttributeValue(amazon.InstanceType, cty.StringVal(pool.InstanceType))
		nodePoolsBlockBody.SetAttributeValue(amazon.DesiredSize, cty.NumberIntVal(pool.DesiredSize))
		nodePoolsBlockBody.SetAttributeValue(amazon.MaxSize, cty.NumberIntVal(pool.MaxSize))
		nodePoolsBlockBody.SetAttributeValue(amazon.MinSize, cty.NumberIntVal(pool.MinSize))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write EKS configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
