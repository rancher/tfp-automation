package hosted

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/google"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetGKE is a function that will set the GKE configurations in the main.tf file.
func SetGKE(terraformConfig *config.TerraformConfig, k8sVersion string, nodePools []config.Nodepool, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	cloudCredBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.CloudCredential, defaults.CloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	googleCredConfigBlock := cloudCredBlockBody.AppendNewBlock(google.GoogleCredentialConfig, nil)
	googleCredConfigBlock.Body().SetAttributeValue(google.AuthEncodedJSON, cty.StringVal(terraformConfig.GoogleCredentials.AuthEncodedJSON))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Cluster, defaults.Cluster})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	gkeConfigBlock := clusterBlockBody.AppendNewBlock(google.GKEConfig, nil)
	gkeConfigBlockBody := gkeConfigBlock.Body()

	gkeConfigBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	cloudCredSecret := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.CloudCredential + "." + defaults.CloudCredential + ".id")},
	}

	gkeConfigBlockBody.SetAttributeRaw(google.GoogleCredentialSecret, cloudCredSecret)
	gkeConfigBlockBody.SetAttributeValue(defaults.Region, cty.StringVal(terraformConfig.GoogleConfig.Region))
	gkeConfigBlockBody.SetAttributeValue(google.ProjectID, cty.StringVal(terraformConfig.GoogleConfig.ProjectID))
	gkeConfigBlockBody.SetAttributeValue(defaults.KubernetesVersion, cty.StringVal(k8sVersion))
	gkeConfigBlockBody.SetAttributeValue(google.Network, cty.StringVal(terraformConfig.GoogleConfig.Network))
	gkeConfigBlockBody.SetAttributeValue(google.Subnetwork, cty.StringVal(terraformConfig.GoogleConfig.Subnetwork))

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := resources.SetResourceNodepoolValidation(terraformConfig, pool, poolNum)
		if err != nil {
			return nil, nil, err
		}

		nodePoolsBlock := gkeConfigBlockBody.AppendNewBlock(google.NodePools, nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeValue(google.InitialNodeCount, cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue(google.MaxPodsConstraint, cty.NumberIntVal(pool.MaxPodsContraint))
		nodePoolsBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+`-pool`+poolNum))
		nodePoolsBlockBody.SetAttributeValue(google.Version, cty.StringVal(k8sVersion))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write GKE configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}
