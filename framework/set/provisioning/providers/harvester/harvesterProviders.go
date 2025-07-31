package harvester

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/harvester"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetHarvesterRKE1Provider is a helper function that will set the Harvester RKE1 terraform configurations in the main.tf file.
func SetHarvesterRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	nodeTemplateBlockBody.SetAttributeValue("engine_install_url", cty.StringVal("https://releases.rancher.com/install-docker/26.1.sh"))

	cloudCredID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.CloudCredential + "." + terraformConfig.ResourcePrefix + ".id")},
	}
	nodeTemplateBlockBody.SetAttributeRaw(defaults.CloudCredentialID, cloudCredID)

	harvesterConfigBlock := nodeTemplateBlockBody.AppendNewBlock(harvester.HarvesterConfig, nil)
	harvesterConfigBlockBody := harvesterConfigBlock.Body()

	if terraformConfig.HarvesterConfig.UserData == "" {
		harvesterConfigBlockBody.SetAttributeRaw(harvester.UserData, hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOT\n#cloud-config\npackage_update: true\npackages:\n  - qemu-guest-agent\nruncmd:\n  - - systemctl\n    - enable\n    - '--now'\n    - qemu-guest-agent.service\nEOT"},
		}))
	}
	harvesterConfigBlockBody.SetAttributeRaw(harvester.NetworkInfo, constructNetworkInfo(terraformConfig.HarvesterConfig.NetworkNames))
	harvesterConfigBlockBody.SetAttributeRaw(harvester.DiskInfo, constructDiskInfo(terraformConfig.HarvesterConfig.ImageName, terraformConfig.HarvesterConfig.DiskSize))

	harvesterConfigBlockBody.SetAttributeValue(harvester.CPUCount, cty.StringVal(terraformConfig.HarvesterConfig.CPUCount))
	harvesterConfigBlockBody.SetAttributeValue(harvester.MemorySize, cty.StringVal(terraformConfig.HarvesterConfig.MemorySize))
	harvesterConfigBlockBody.SetAttributeValue(harvester.SSHUser, cty.StringVal(terraformConfig.HarvesterConfig.SSHUser))
	harvesterConfigBlockBody.SetAttributeValue(harvester.VMNamespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))
}

// SetHarvesterCredentialProvider is a helper function that will set the Harvester cloud provider in main.tf
func SetHarvesterCredentialProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Rancher2 + "." + defaults.StandardUser)},
	}

	cloudCredBlockBody.SetAttributeRaw(defaults.Provider, provider)
	cloudCredBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

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
