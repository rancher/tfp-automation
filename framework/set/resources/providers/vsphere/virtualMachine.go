package vsphere

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereVirtualMachine is a function that will set the vSphere virtual machine configuration in the main.tf file.
func CreateVsphereVirtualMachine(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	vmBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.VsphereVirutalMachine, hostnamePrefix})
	vmBlockBody := vmBlock.Body()

	if strings.Contains(terraformConfig.Module, defaults.Custom) {
		vmBlockBody.SetAttributeValue(defaults.Count, cty.NumberIntVal(terratestConfig.NodeCount))

		vmNameExpression := fmt.Sprintf(` "%s-${%s.%s}"`, hostnamePrefix, defaults.Count, defaults.Index)
		vmNameValue := hclwrite.Tokens{
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(vmNameExpression)},
		}

		vmBlockBody.SetAttributeRaw(defaults.ResourceName, vmNameValue)
	} else {
		vmBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(hostnamePrefix))
	}

	resourcePoolExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereComputeCluster + `.` + defaults.VsphereComputeCluster + `.` + resourcePoolID)
	resourcePoolValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(resourcePoolExpression)},
	}

	vmBlockBody.SetAttributeRaw(resourcePoolID, resourcePoolValue)

	dataStoreExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereDatastore + `.` + defaults.VsphereDatastore + `.id`)
	dataStoreValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dataStoreExpression)},
	}

	vmBlockBody.SetAttributeRaw(datastoreID, dataStoreValue)
	vmBlockBody.SetAttributeValue(defaults.Folder, cty.StringVal(terraformConfig.VsphereConfig.Folder))

	cpuCount, err := strconv.ParseInt(terraformConfig.VsphereConfig.CPUCount, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid CPU count value: %s", terraformConfig.VsphereConfig.CPUCount))
	}

	vmBlockBody.SetAttributeValue(defaults.NumCPUs, cty.NumberIntVal(cpuCount))

	memory, err := strconv.ParseInt(terraformConfig.VsphereConfig.MemorySize, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid memory size value: %s", terraformConfig.VsphereConfig.MemorySize))
	}

	vmBlockBody.SetAttributeValue(defaults.Memory, cty.NumberIntVal(memory))
	vmBlockBody.SetAttributeValue(guestID, cty.StringVal(terraformConfig.VsphereConfig.GuestID))

	vmBlockBody.AppendNewline()

	cdROMBlock := vmBlockBody.AppendNewBlock(defaults.CDROM, nil)
	cdROMBlockBody := cdROMBlock.Body()

	cdROMBlockBody.SetAttributeValue(clientDevice, cty.BoolVal(true))
	vmBlockBody.AppendNewline()

	networkBlock := vmBlockBody.AppendNewBlock(defaults.NetworkInterface, nil)
	networkBlockBody := networkBlock.Body()

	networkExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereNetwork + `.` + defaults.VsphereNetwork + `.id`)
	networkValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(networkExpression)},
	}

	networkBlockBody.SetAttributeRaw(defaults.NetworkID, networkValue)
	vmBlockBody.AppendNewline()

	diskBlock := vmBlockBody.AppendNewBlock(defaults.Disk, nil)
	diskBlockBody := diskBlock.Body()

	if strings.Contains(terraformConfig.Module, defaults.Custom) {
		diskSizeExpression := fmt.Sprintf(` "%s-${%s.%s}"`, hostnamePrefix, defaults.Count, defaults.Index)
		diskSizeValue := hclwrite.Tokens{
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(diskSizeExpression)},
		}

		diskBlockBody.SetAttributeRaw(defaults.Label, diskSizeValue)
	} else {
		diskBlockBody.SetAttributeValue(defaults.Label, cty.StringVal(hostnamePrefix))
	}

	diskSize, err := strconv.ParseInt(terraformConfig.VsphereConfig.DiskSize, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid disk size value: %s", terraformConfig.VsphereConfig.DiskSize))
	}

	diskBlockBody.SetAttributeValue(defaults.Size, cty.NumberIntVal(diskSize))
	vmBlockBody.AppendNewline()

	cloneBlock := vmBlockBody.AppendNewBlock(clone, nil)
	cloneBlockBody := cloneBlock.Body()

	templateUUIDExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereVirutalMachine + `.` + defaults.VsphereVirtualMachineTemplate + `.id`)
	templateUUIDValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(templateUUIDExpression)},
	}

	cloneBlockBody.SetAttributeRaw(templateUUID, templateUUIDValue)
	vmBlockBody.AppendNewline()

	extraConfigBlock := vmBlockBody.AppendNewBlock(defaults.ExtraConfig+" =", nil)
	extraConfigBlockBody := extraConfigBlock.Body()

	extraConfigBlockBody.SetAttributeValue(diskEnableUUID, cty.BoolVal(true))
}
