package azure

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/azure"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// SetAzureRKE2K3SProvider is a helper function that will set the Azure RKE2/K3S Terraform provider details in the main.tf file.
func SetAzureRKE2K3SProvider(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	cloudCredBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.CloudCredential, terraformConfig.ResourcePrefix})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	azureCredBlock := cloudCredBlockBody.AppendNewBlock(azure.AzureCredentialConfig, nil)
	azureCredBlockBody := azureCredBlock.Body()

	azureCredBlockBody.SetAttributeValue(azure.ClientID, cty.StringVal(terraformConfig.AzureCredentials.ClientID))
	azureCredBlockBody.SetAttributeValue(azure.ClientSecret, cty.StringVal(terraformConfig.AzureCredentials.ClientSecret))
	azureCredBlockBody.SetAttributeValue(azure.SubscriptionID, cty.StringVal(terraformConfig.AzureCredentials.SubscriptionID))
	azureCredBlockBody.SetAttributeValue(azure.Environment, cty.StringVal(terraformConfig.AzureCredentials.Environment))
	azureCredBlockBody.SetAttributeValue(azure.TenantID, cty.StringVal(terraformConfig.AzureCredentials.TenantID))

}
