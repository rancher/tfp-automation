package harvester

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/harvester"
	"github.com/zclconf/go-cty/cty"
)

// SetHarvesterRKE2K3SMachineConfig is a helper function that will set the Harvester RKE2/K3S terraform machine configurations in the main.tf file.
func SetHarvesterRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	harvesterConfigBlock := machineConfigBlockBody.AppendNewBlock(harvester.HarvesterConfig, nil)
	harvesterConfigBlockBody := harvesterConfigBlock.Body()

	harvesterConfigBlockBody.SetAttributeRaw(harvester.NetworkInfo, constructNetworkInfo(terraformConfig.HarvesterConfig.NetworkNames))
	harvesterConfigBlockBody.SetAttributeRaw(harvester.DiskInfo, constructDiskInfo(terraformConfig.HarvesterConfig.ImageName, terraformConfig.HarvesterConfig.DiskSize))

	if terraformConfig.HarvesterConfig.UserData == "" {
		harvesterConfigBlockBody.SetAttributeRaw(harvester.UserData, hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOT\n#cloud-config\npackage_update: true\npackages:\n  - qemu-guest-agent\nruncmd:\n  - - systemctl\n    - enable\n    - '--now'\n    - qemu-guest-agent.service\nEOT"},
		}))
	}

	harvesterConfigBlockBody.SetAttributeValue(harvester.CPUCount, cty.StringVal(terraformConfig.HarvesterConfig.CPUCount))
	harvesterConfigBlockBody.SetAttributeValue(harvester.MemorySize, cty.StringVal(terraformConfig.HarvesterConfig.MemorySize))
	harvesterConfigBlockBody.SetAttributeValue(harvester.SSHUser, cty.StringVal(terraformConfig.HarvesterConfig.SSHUser))
	harvesterConfigBlockBody.SetAttributeValue(harvester.VMNamespace, cty.StringVal(terraformConfig.HarvesterConfig.VMNamespace))
}
