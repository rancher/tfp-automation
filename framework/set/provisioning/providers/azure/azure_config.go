package azure

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// SetAzureRKE2K3SMachineConfig is a helper function that will set the Azure RKE2/K3S Terraform machine configurations in the main.tf file.
func SetAzureRKE2K3SMachineConfig(machineConfigBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	azureConfigBlock := machineConfigBlockBody.AppendNewBlock("azure_config", nil)
	azureConfigBlockBody := azureConfigBlock.Body()

	openPorts := make([]cty.Value, len(terraformConfig.AzureConfig.OpenPort))
	for i, port := range terraformConfig.AzureConfig.OpenPort {
		openPorts[i] = cty.StringVal(port)
	}

	azureConfigBlockBody.SetAttributeValue("availability_set", cty.StringVal(terraformConfig.AzureConfig.AvailabilitySet))
	azureConfigBlockBody.SetAttributeValue("custom_data", cty.StringVal(terraformConfig.AzureConfig.CustomData))
	azureConfigBlockBody.SetAttributeValue("disk_size", cty.StringVal(terraformConfig.AzureConfig.DiskSize))
	azureConfigBlockBody.SetAttributeValue("fault_domain_count", cty.StringVal(terraformConfig.AzureConfig.FaultDomainCount))
	azureConfigBlockBody.SetAttributeValue("image", cty.StringVal(terraformConfig.AzureConfig.Image))
	azureConfigBlockBody.SetAttributeValue("location", cty.StringVal(terraformConfig.AzureConfig.Location))
	azureConfigBlockBody.SetAttributeValue("managed_disks", cty.BoolVal(terraformConfig.AzureConfig.ManagedDisks))
	azureConfigBlockBody.SetAttributeValue("no_public_ip", cty.BoolVal(terraformConfig.AzureConfig.NoPublicIP))
	azureConfigBlockBody.SetAttributeValue("open_port", cty.ListVal(openPorts))
	azureConfigBlockBody.SetAttributeValue("private_ip_address", cty.StringVal(terraformConfig.AzureConfig.PrivateIPAddress))
	azureConfigBlockBody.SetAttributeValue("resource_group", cty.StringVal(terraformConfig.AzureConfig.ResourceGroup))
	azureConfigBlockBody.SetAttributeValue("size", cty.StringVal(terraformConfig.AzureConfig.Size))
	azureConfigBlockBody.SetAttributeValue("ssh_user", cty.StringVal(terraformConfig.AzureConfig.SSHUser))
	azureConfigBlockBody.SetAttributeValue("static_public_ip", cty.BoolVal(terraformConfig.AzureConfig.StaticPublicIP))
	azureConfigBlockBody.SetAttributeValue("storage_type", cty.StringVal(terraformConfig.AzureConfig.StorageType))
	azureConfigBlockBody.SetAttributeValue("tenant_id", cty.StringVal(terraformConfig.AzureConfig.TenantID))
	azureConfigBlockBody.SetAttributeValue("update_domain_count", cty.StringVal(terraformConfig.AzureConfig.UpdateDomainCount))
	azureConfigBlockBody.SetAttributeValue("use_private_ip", cty.BoolVal(terraformConfig.AzureConfig.UsePrivateIP))
}
