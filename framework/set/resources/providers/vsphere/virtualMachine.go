package vsphere

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

// CreateVsphereVirtualMachine is a function that will set the vSphere virtual machine configuration in the main.tf file.
func CreateVsphereVirtualMachine(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	hostnamePrefix string) {
	vmBlock := rootBody.AppendNewBlock(general.Resource, []string{vsphere.VsphereVirtualMachine, hostnamePrefix})
	vmBlockBody := vmBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		totalNodeCount := terratestConfig.EtcdCount + terratestConfig.ControlPlaneCount + terratestConfig.WorkerCount
		vmBlockBody.SetAttributeValue(general.Count, cty.NumberIntVal(totalNodeCount))

		vmNameExpression := fmt.Sprintf(` "%s-${%s.%s}"`, hostnamePrefix, general.Count, general.Index)
		vmNameValue := hclwrite.Tokens{
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(vmNameExpression)},
		}

		vmBlockBody.SetAttributeRaw(general.ResourceName, vmNameValue)
	} else {
		vmBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(hostnamePrefix))
	}

	resourcePoolExpression := fmt.Sprintf(general.Data + `.` + vsphere.VsphereComputeCluster + `.` + vsphere.VsphereComputeCluster + `.` + resourcePoolID)
	resourcePoolValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(resourcePoolExpression)},
	}

	vmBlockBody.SetAttributeRaw(resourcePoolID, resourcePoolValue)

	dataStoreExpression := fmt.Sprintf(general.Data + `.` + vsphere.VsphereDatastore + `.` + vsphere.VsphereDatastore + `.id`)
	dataStoreValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dataStoreExpression)},
	}

	vmBlockBody.SetAttributeRaw(datastoreID, dataStoreValue)
	vmBlockBody.SetAttributeValue(vsphere.Folder, cty.StringVal(terraformConfig.VsphereConfig.Folder))

	cpuCount, err := strconv.ParseInt(terraformConfig.VsphereConfig.CPUCount, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid CPU count value: %s", terraformConfig.VsphereConfig.CPUCount))
	}

	vmBlockBody.SetAttributeValue(vsphere.NumCPUs, cty.NumberIntVal(cpuCount))

	memory, err := strconv.ParseInt(terraformConfig.VsphereConfig.MemorySize, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid memory size value: %s", terraformConfig.VsphereConfig.MemorySize))
	}

	vmBlockBody.SetAttributeValue(vsphere.Memory, cty.NumberIntVal(memory))
	vmBlockBody.SetAttributeValue(guestID, cty.StringVal(terraformConfig.VsphereConfig.GuestID))

	vmBlockBody.AppendNewline()

	cdROMBlock := vmBlockBody.AppendNewBlock(vsphere.CDROM, nil)
	cdROMBlockBody := cdROMBlock.Body()

	cdROMBlockBody.SetAttributeValue(clientDevice, cty.BoolVal(true))
	vmBlockBody.AppendNewline()

	vappBlock := vmBlockBody.AppendNewBlock(vsphere.Vapp, nil)
	vappBlockBody := vappBlock.Body()

	propertiesBlock := vappBlockBody.AppendNewBlock(vsphere.VappProperties+" =", nil)
	propertiesBlockBody := propertiesBlock.Body()

	propertiesBlockBody.SetAttributeValue(publicKeys, cty.StringVal(terraformConfig.PrivateKeyPath))

	networkBlock := vmBlockBody.AppendNewBlock(vsphere.NetworkInterface, nil)
	networkBlockBody := networkBlock.Body()

	networkExpression := fmt.Sprintf(general.Data + `.` + vsphere.VsphereNetwork + `.` + vsphere.VsphereNetwork + `.id`)
	networkValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(networkExpression)},
	}

	networkBlockBody.SetAttributeRaw(vsphere.NetworkID, networkValue)
	vmBlockBody.AppendNewline()

	diskBlock := vmBlockBody.AppendNewBlock(vsphere.Disk, nil)
	diskBlockBody := diskBlock.Body()

	if strings.Contains(terraformConfig.Module, general.Custom) {
		diskSizeExpression := fmt.Sprintf(` "%s-${%s.%s}"`, hostnamePrefix, general.Count, general.Index)
		diskSizeValue := hclwrite.Tokens{
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(diskSizeExpression)},
		}

		diskBlockBody.SetAttributeRaw(general.Label, diskSizeValue)
	} else {
		diskBlockBody.SetAttributeValue(general.Label, cty.StringVal(hostnamePrefix))
	}

	diskSize, err := strconv.ParseInt(terraformConfig.VsphereConfig.DiskSize, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Invalid disk size value: %s", terraformConfig.VsphereConfig.DiskSize))
	}

	diskBlockBody.SetAttributeValue(vsphere.Size, cty.NumberIntVal(diskSize))
	vmBlockBody.AppendNewline()

	cloneBlock := vmBlockBody.AppendNewBlock(clone, nil)
	cloneBlockBody := cloneBlock.Body()

	templateUUIDExpression := fmt.Sprintf(general.Data + `.` + vsphere.VsphereVirtualMachine + `.` + vsphere.VsphereVirtualMachineTemplate + `.id`)
	templateUUIDValue := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(templateUUIDExpression)},
	}

	cloneBlockBody.SetAttributeRaw(templateUUID, templateUUIDValue)
	vmBlockBody.AppendNewline()

	extraConfigBlock := vmBlockBody.AppendNewBlock(vsphere.ExtraConfig+" =", nil)
	extraConfigBlockBody := extraConfigBlock.Body()

	extraConfigBlockBody.SetAttributeValue(diskEnableUUID, cty.BoolVal(true))
}
