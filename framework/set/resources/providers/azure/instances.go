package azure

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

// CreateAzureInstances is a function that will set the Azure VM instances configurations in the main.tf file.
func CreateAzureInstances(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, hostnamePrefix string) {
	vmBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureLinuxVirtualMachine, hostnamePrefix})
	vmBlockBody := vmBlock.Body()

	vmBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-"+hostnamePrefix))

	expression := azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".name"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	vmBlockBody.SetAttributeRaw(resourceGroupName, values)

	expression = azure.AzureResourceGroup + "." + azure.AzureResourceGroup + ".location"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	vmBlockBody.SetAttributeRaw(location, values)
	vmBlockBody.SetAttributeValue(general.Size, cty.StringVal(terraformConfig.AzureConfig.Size))
	vmBlockBody.SetAttributeValue(adminUsername, cty.StringVal(terraformConfig.AzureConfig.SSHUser))

	expression = "[" + azure.AzureNetworkInterface + "." + azure.AzureNetworkInterface + "-" + hostnamePrefix + ".id" + "]"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	vmBlockBody.SetAttributeRaw(networkInterfaceIDs, values)

	sshKeyBlock := vmBlockBody.AppendNewBlock(azure.AdminSSHKey, nil)
	sshKeyBlockBody := sshKeyBlock.Body()

	sshKeyBlockBody.SetAttributeValue(general.Username, cty.StringVal(terraformConfig.AzureConfig.SSHUser))

	expression = fmt.Sprintf(`"${file("%s")}"`, terraformConfig.AzureConfig.KeyPath)
	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	sshKeyBlockBody.SetAttributeRaw(general.PublicKey, value)

	osDiskBlock := vmBlockBody.AppendNewBlock(azure.OSDisk, nil)
	osDiskBlockBody := osDiskBlock.Body()

	osDiskBlockBody.SetAttributeValue(caching, cty.StringVal("ReadWrite"))
	osDiskBlockBody.SetAttributeValue(storageAccountType, cty.StringVal("Standard_LRS"))
	osDiskBlockBody.SetAttributeValue(diskSizeGB, cty.StringVal(terraformConfig.AzureConfig.DiskSize))

	sourceImageRefBlock := vmBlockBody.AppendNewBlock(sourceImageReference, nil)
	sourceImageRefBlockBody := sourceImageRefBlock.Body()

	sourceImageRefBlockBody.SetAttributeValue(publisher, cty.StringVal(terraformConfig.AzureConfig.Publisher))
	sourceImageRefBlockBody.SetAttributeValue(offer, cty.StringVal(terraformConfig.AzureConfig.ImageOffer))
	sourceImageRefBlockBody.SetAttributeValue(sku, cty.StringVal(terraformConfig.AzureConfig.SKU))
	sourceImageRefBlockBody.SetAttributeValue(version, cty.StringVal(terraformConfig.AzureConfig.ImageVersion))
}
