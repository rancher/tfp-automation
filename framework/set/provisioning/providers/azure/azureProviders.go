package azure

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/azure"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// SetAzureRKE1Provider is a helper function that will set the Azure RKE1
// Terraform configurations in the main.tf file.
func SetAzureRKE1Provider(nodeTemplateBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	azureConfigBlock := nodeTemplateBlockBody.AppendNewBlock(azure.AzureConfig, nil)
	azureConfigBlockBody := azureConfigBlock.Body()

	openPorts := make([]cty.Value, len(terraformConfig.AzureConfig.OpenPort))
	for i, port := range terraformConfig.AzureConfig.OpenPort {
		openPorts[i] = cty.StringVal(port)
	}

	azureConfigBlockBody.SetAttributeValue(azure.AvailabilitySet, cty.StringVal(terraformConfig.AzureConfig.AvailabilitySet))
	azureConfigBlockBody.SetAttributeValue(azure.ClientID, cty.StringVal(terraformConfig.AzureCredentials.ClientID))
	azureConfigBlockBody.SetAttributeValue(azure.ClientSecret, cty.StringVal(terraformConfig.AzureCredentials.ClientSecret))
	azureConfigBlockBody.SetAttributeValue(azure.SubscriptionID, cty.StringVal(terraformConfig.AzureCredentials.SubscriptionID))
	azureConfigBlockBody.SetAttributeValue(azure.Environment, cty.StringVal(terraformConfig.AzureCredentials.Environment))
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

// SetAzureRKE2K3SProvider is a helper function that will set the Azure RKE2/K3S Terraform provider details in the main.tf file.
func SetAzureRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	azureCredBlock := cloudCredBlockBody.AppendNewBlock(azure.AzureCredentialConfig, nil)
	azureCredBlockBody := azureCredBlock.Body()

	azureCredBlockBody.SetAttributeValue(azure.ClientID, cty.StringVal(terraformConfig.AzureCredentials.ClientID))
	azureCredBlockBody.SetAttributeValue(azure.ClientSecret, cty.StringVal(terraformConfig.AzureCredentials.ClientSecret))
	azureCredBlockBody.SetAttributeValue(azure.SubscriptionID, cty.StringVal(terraformConfig.AzureCredentials.SubscriptionID))
	azureCredBlockBody.SetAttributeValue(azure.Environment, cty.StringVal(terraformConfig.AzureCredentials.Environment))
	azureCredBlockBody.SetAttributeValue(azure.TenantID, cty.StringVal(terraformConfig.AzureCredentials.TenantID))

}
