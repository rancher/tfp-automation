package provisioning

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/google"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetGKE is a function that will set the GKE configurations in the main.tf file.
func SetGKE(clusterName, k8sVersion string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	framework.LoadConfig(configs.Rancher, rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig(configs.Terraform, terraformConfig)

	newFile, rootBody := SetProvidersAndUsersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock(resource, []string{cloudCredential, cloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.CloudCredentialName))

	googleCredConfigBlock := cloudCredBlockBody.AppendNewBlock(google.GoogleCredentialConfig, nil)
	googleCredConfigBlock.Body().SetAttributeValue(google.AuthEncodedJSON, cty.StringVal(terraformConfig.GoogleConfig.AuthEncodedJSON))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(resource, []string{cluster, cluster})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(resourceName, cty.StringVal(clusterName))

	gkeConfigBlock := clusterBlockBody.AppendNewBlock(google.GKEConfig, nil)
	gkeConfigBlockBody := gkeConfigBlock.Body()

	gkeConfigBlockBody.SetAttributeValue(resourceName, cty.StringVal(clusterName))

	cloudCredSecret := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(cloudCredential + "." + cloudCredential + ".id")},
	}

	gkeConfigBlockBody.SetAttributeRaw(google.GoogleCredentialSecret, cloudCredSecret)
	gkeConfigBlockBody.SetAttributeValue(region, cty.StringVal(terraformConfig.GoogleConfig.Region))
	gkeConfigBlockBody.SetAttributeValue(google.ProjectID, cty.StringVal(terraformConfig.GoogleConfig.ProjectID))
	gkeConfigBlockBody.SetAttributeValue(kubernetesVersion, cty.StringVal(k8sVersion))
	gkeConfigBlockBody.SetAttributeValue(google.Network, cty.StringVal(terraformConfig.GoogleConfig.Network))
	gkeConfigBlockBody.SetAttributeValue(google.Subnetwork, cty.StringVal(terraformConfig.GoogleConfig.Subnetwork))

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolsBlock := gkeConfigBlockBody.AppendNewBlock(google.NodePools, nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeValue(google.InitialNodeCount, cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue(google.MaxPodsConstraint, cty.NumberIntVal(pool.MaxPodsContraint))
		nodePoolsBlockBody.SetAttributeValue(resourceName, cty.StringVal(terraformConfig.HostnamePrefix+`-pool`+poolNum))
		nodePoolsBlockBody.SetAttributeValue(google.Version, cty.StringVal(k8sVersion))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write GKE configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
