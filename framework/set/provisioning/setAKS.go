package provisioning

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	framework "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/tfp-automation/config"
	format "github.com/rancher/tfp-automation/framework/format"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// SetAKS is a function that will set the AKS configurations in the main.tf file.
func SetAKS(clusterName, k8sVersion string, nodePools []config.Nodepool, file *os.File) error {
	rancherConfig := new(rancher.Config)
	framework.LoadConfig("rancher", rancherConfig)

	terraformConfig := new(config.TerraformConfig)
	framework.LoadConfig("terraform", terraformConfig)

	newFile, rootBody := SetProvidersTF(rancherConfig, terraformConfig)

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cloud_credential", "rancher2_cloud_credential"})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue("name", cty.StringVal(terraformConfig.CloudCredentialName))

	azCredConfigBlock := cloudCredBlockBody.AppendNewBlock("azure_credential_config", nil)
	azCredConfigBlockBody := azCredConfigBlock.Body()

	azCredConfigBlockBody.SetAttributeValue("client_id", cty.StringVal(terraformConfig.AzureConfig.ClientID))
	azCredConfigBlockBody.SetAttributeValue("client_secret", cty.StringVal(terraformConfig.AzureConfig.ClientSecret))
	azCredConfigBlockBody.SetAttributeValue("subscription_id", cty.StringVal(terraformConfig.AzureConfig.SubscriptionID))

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_cluster", "rancher2_cluster"})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue("name", cty.StringVal(clusterName))

	aksConfigBlock := clusterBlockBody.AppendNewBlock("aks_config_v2", nil)
	aksConfigBlockBody := aksConfigBlock.Body()

	cloudCredID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`rancher2_cloud_credential.rancher2_cloud_credential.id`)},
	}

	aksConfigBlockBody.SetAttributeRaw("cloud_credential_id", cloudCredID)
	aksConfigBlockBody.SetAttributeValue("resource_group", cty.StringVal(terraformConfig.AzureConfig.ResourceGroup))
	aksConfigBlockBody.SetAttributeValue("resource_location", cty.StringVal(terraformConfig.AzureConfig.ResourceLocation))
	aksConfigBlockBody.SetAttributeValue("dns_prefix", cty.StringVal(terraformConfig.HostnamePrefix))
	aksConfigBlockBody.SetAttributeValue("kubernetes_version", cty.StringVal(k8sVersion))
	aksConfigBlockBody.SetAttributeValue("network_plugin", cty.StringVal(terraformConfig.NetworkPlugin))

	availabilityZones := format.ListOfStrings(terraformConfig.AzureConfig.AvailabilityZones)

	for count, pool := range nodePools {
		poolNum := strconv.Itoa(count)

		_, err := SetResourceNodepoolValidation(pool, poolNum)
		if err != nil {
			return err
		}

		nodePoolsBlock := aksConfigBlockBody.AppendNewBlock("node_pools", nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeRaw("availability_zones", availabilityZones)
		nodePoolsBlockBody.SetAttributeValue("name", cty.StringVal("pool"+poolNum))
		nodePoolsBlockBody.SetAttributeValue("count", cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue("orchestrator_version", cty.StringVal(k8sVersion))
		nodePoolsBlockBody.SetAttributeValue("os_disk_size_gb", cty.NumberIntVal(terraformConfig.AzureConfig.OSDiskSizeGB))
		nodePoolsBlockBody.SetAttributeValue("vm_size", cty.StringVal(terraformConfig.AzureConfig.VMSize))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write AKS configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
