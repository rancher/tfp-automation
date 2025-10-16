package vsphere

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	locals             = "locals"
	requiredProviders  = "required_providers"
	serverOne          = "server1"
	serverTwo          = "server2"
	serverThree        = "server3"
	instanceIDs        = "instance_ids"
	allowUnverifiedSSL = "allow_unverified_ssl"
)

// CreateVsphereTerraformProviderBlock will up the terraform block with the required vsphere provider.
func CreateVsphereTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	cloudProviderVersion := os.Getenv("CLOUD_PROVIDER_VERSION")
	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Vsphere, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.VsphereSource),
		defaults.Version: cty.StringVal(cloudProviderVersion),
	}))
}

// CreateVsphereProviderBlock will set up the vsphere provider block.
func CreateVsphereProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	vsphereProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Vsphere})
	vsphereProvBlockBody := vsphereProvBlock.Body()

	vsphereProvBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.VsphereCredentials.Username))
	vsphereProvBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.VsphereCredentials.Password))
	vsphereProvBlockBody.SetAttributeValue(defaults.VsphereServer, cty.StringVal(terraformConfig.VsphereCredentials.Vcenter))
	vsphereProvBlockBody.SetAttributeValue(allowUnverifiedSSL, cty.BoolVal(true))
}

// CreateVsphereLocalBlock will set up the local block. Returns the local block.
func CreateVsphereLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	instanceIds := map[string]any{
		serverOne:   defaults.VsphereVirtualMachine + "." + serverOne + ".id",
		serverTwo:   defaults.VsphereVirtualMachine + "." + serverTwo + ".id",
		serverThree: defaults.VsphereVirtualMachine + "." + serverThree + ".id",
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
