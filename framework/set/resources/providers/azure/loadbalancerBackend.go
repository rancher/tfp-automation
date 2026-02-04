package azure

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/zclconf/go-cty/cty"
)

const (
	backendAddressPoolIDs = "backend_address_pool_ids"
	backendPort           = "backend_port"
	frontendPort          = "frontend_port"
	frontendIPConfig      = "frontend_ip_configuration_name"
	loadBalancerID        = "loadbalancer_id"
	internalInSeconds     = "interval_in_seconds"
	numberOfProbes        = "number_of_probes"
	probeID               = "probe_id"
)

// CreateAzureLoadBalancerBackend is a function that will set the load balancer backend configurations in the main.tf file.
func CreateAzureLoadBalancerBackend(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	lbBlock := rootBody.AppendNewBlock(general.Resource, []string{azure.AzureLoadBalancerBackendAddressPool, azure.AzureLoadBalancerBackendAddressPool})
	lbBlockBody := lbBlock.Body()

	lbBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+"-lb-backend"))

	expression := azure.AzureLoadBalancer + "." + azure.AzureLoadBalancer + ".id"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	lbBlockBody.SetAttributeRaw(loadBalancerID, values)
}
