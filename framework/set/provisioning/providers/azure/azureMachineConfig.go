package azure

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/azure"
	"github.com/zclconf/go-cty/cty"
)

// SetAzureRKE2K3SMachineConfig is a helper function that will set the Azure RKE2/K3S
// Terraform machine configurations in the main.tf file.
func SetAzureRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	azureConfigBlock := machineConfigBlockBody.AppendNewBlock(azure.AzureConfig, nil)
	azureConfigBlockBody := azureConfigBlock.Body()

	openPorts := make([]cty.Value, len(terraformConfig.AzureConfig.OpenPort))
	for i, port := range terraformConfig.AzureConfig.OpenPort {
		openPorts[i] = cty.StringVal(port)
	}

	azureConfigBlockBody.SetAttributeValue(azure.AvailabilitySet, cty.StringVal(terraformConfig.AzureConfig.AvailabilitySet))
	azureConfigBlockBody.SetAttributeValue(azure.CustomData, cty.StringVal(terraformConfig.AzureConfig.CustomData))
	azureConfigBlockBody.SetAttributeValue(azure.DiskSize, cty.StringVal(terraformConfig.AzureConfig.DiskSize))
	azureConfigBlockBody.SetAttributeValue(azure.FaultDomainCount, cty.StringVal(terraformConfig.AzureConfig.FaultDomainCount))
	azureConfigBlockBody.SetAttributeValue(azure.Image, cty.StringVal(terraformConfig.AzureConfig.Image))
	azureConfigBlockBody.SetAttributeValue(azure.Location, cty.StringVal(terraformConfig.AzureConfig.Location))
	azureConfigBlockBody.SetAttributeValue(azure.ManagedDisks, cty.BoolVal(terraformConfig.AzureConfig.ManagedDisks))
	azureConfigBlockBody.SetAttributeValue(azure.NoPublicIP, cty.BoolVal(terraformConfig.AzureConfig.NoPublicIP))
	azureConfigBlockBody.SetAttributeValue(azure.OpenPort, cty.ListVal(openPorts))
	azureConfigBlockBody.SetAttributeValue(azure.PrivateIPAddress, cty.StringVal(terraformConfig.AzureConfig.PrivateIPAddress))
	azureConfigBlockBody.SetAttributeValue(azure.ResourceGroup, cty.StringVal(terraformConfig.AzureConfig.ResourceGroup))
	azureConfigBlockBody.SetAttributeValue(azure.Size, cty.StringVal(terraformConfig.AzureConfig.Size))
	azureConfigBlockBody.SetAttributeValue(azure.SSHUser, cty.StringVal(terraformConfig.AzureConfig.SSHUser))
	azureConfigBlockBody.SetAttributeValue(azure.StaticPublicIP, cty.BoolVal(terraformConfig.AzureConfig.StaticPublicIP))
	azureConfigBlockBody.SetAttributeValue(azure.StorageType, cty.StringVal(terraformConfig.AzureConfig.StorageType))
	azureConfigBlockBody.SetAttributeValue(azure.UpdateDomainCount, cty.StringVal(terraformConfig.AzureConfig.UpdateDomainCount))
	azureConfigBlockBody.SetAttributeValue(azure.UsePrivateIP, cty.BoolVal(terraformConfig.AzureConfig.UsePrivateIP))
}
