package azure

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

const (
	adminUsername              = "admin_username"
	addressPrefixes            = "address_prefixes"
	addressSpace               = "address_space"
	allocationMethod           = "allocation_method"
	backendAddressPoolID       = "backend_address_pool_id"
	caching                    = "caching"
	diskSizeGB                 = "disk_size_gb"
	domainNameLabel            = "domain_name_label"
	frontend                   = "frontend"
	instanceIDs                = "instance_ids"
	ipConfiguration            = "ip_configuration"
	ipConfigurationName        = "ip_configuration_name"
	internal                   = "internal"
	loadBalancerIP             = "load_balancer_ip"
	locals                     = "locals"
	location                   = "location"
	networkInterfaceID         = "network_interface_id"
	networkInterfaceIDs        = "network_interface_ids"
	networkSecurityGroupID     = "network_security_group_id"
	none                       = "none"
	offer                      = "offer"
	privateIPAddressAllocation = "private_ip_address_allocation"
	publicIPAddressID          = "public_ip_address_id"
	publisher                  = "publisher"
	resourceGroupName          = "resource_group_name"
	requiredProviders          = "required_providers"
	securityRule               = "security_rule"
	serverOne                  = "server1"
	serverTwo                  = "server2"
	serverThree                = "server3"
	sku                        = "sku"
	sourceImageReference       = "source_image_reference"
	storageAccountType         = "storage_account_type"
	subnetID                   = "subnet_id"
	version                    = "version"
	virtualNetworkName         = "virtual_network_name"
)

// CreateAzureTerraformProviderBlock will up the terraform block with the required azure provider.
func CreateAzureTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(azure.AzureRM, cty.ObjectVal(map[string]cty.Value{
		general.Source:  cty.StringVal(azure.AzureSource),
		general.Version: cty.StringVal("=" + cloudProviderVersion),
	}))
}

// CreateAzureProviderBlock will set up the azure provider block.
func CreateAzureProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	azureProvBlock := rootBody.AppendNewBlock(general.Provider, []string{azure.AzureRM})
	azureProvBlockBody := azureProvBlock.Body()

	azureProvBlockBody.SetAttributeValue(azure.SubscriptionID, cty.StringVal(terraformConfig.AzureCredentials.SubscriptionID))
	azureProvBlockBody.SetAttributeValue(azure.ResourceProviderRegistrations, cty.StringVal(none))

	featuresBlock := azureProvBlockBody.AppendNewBlock(azure.Features, nil)
	_ = featuresBlock.Body()
}

// CreateAzureLocalBlock will set up the local block. Returns the local block.
func CreateAzureLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	var instanceIds map[string]any

	if terraformConfig.Provider == providers.AKS {
		instanceIds = map[string]any{
			serverOne: azure.AzureLinuxVirtualMachine + "." + serverOne + ".id",
		}
	} else {
		instanceIds = map[string]any{
			serverOne:   azure.AzureLinuxVirtualMachine + "." + serverOne + ".id",
			serverTwo:   azure.AzureLinuxVirtualMachine + "." + serverTwo + ".id",
			serverThree: azure.AzureLinuxVirtualMachine + "." + serverThree + ".id",
		}
	}

	instanceIdsBlock := localBlockBody.AppendNewBlock(instanceIDs+" =", nil)
	instanceIdsBlockBody := instanceIdsBlock.Body()

	for key, value := range instanceIds {
		expression := value.(string)
		instanceValues := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		instanceIdsBlockBody.SetAttributeRaw(key, instanceValues)
	}
}
