package harvester

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/harvester"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// SetHarvesterCredentialProvider is a helper function that will set the Harvester cloud provider in main.tf
func SetHarvesterCredentialProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	harvesterCredBlock := cloudCredBlockBody.AppendNewBlock(harvester.HarvesterCredentialConfig, nil)
	harvesterCredBlockBody := harvesterCredBlock.Body()

	harvesterCredBlockBody.SetAttributeValue(harvester.ClusterID, cty.StringVal(terraformConfig.HarvesterCredentials.ClusterID))
	harvesterCredBlockBody.SetAttributeValue(harvester.ClusterType, cty.StringVal(terraformConfig.HarvesterCredentials.ClusterType))
	harvesterCredBlockBody.SetAttributeValue(harvester.KubeconfigContent, cty.StringVal(terraformConfig.HarvesterCredentials.KubeconfigContent))
}

func constructNetworkInfo(networkNames []string) hclwrite.Tokens {
	var netString string
	for _, net := range networkNames {
		netString += "{\n\t\t\"networkName\": \"" + net + "\"\n\t},"
	}
	netString = netString[:len(netString)-1]

	return hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\n{\n\t\"interfaces\": [" + netString + "]\n}" + "\nEOF"},
	})
}

func constructDiskInfo(imageName, diskSize string) hclwrite.Tokens {
	var diskString = "{\n\t\t\"imageName\": \"" + imageName + "\",\n\t"
	diskString += "\t\"size\": " + diskSize + ",\n\t"
	diskString += "\t\"bootOrder\": 1 \n\t"
	diskString += "}"

	return hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: "<<EOF\n{\n\t\"disks\": [" + diskString + "]\n}" + "\nEOF"},
	})
}
