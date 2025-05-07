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
	k3sServerOne       = "k3s_server1"
	k3sServerTwo       = "k3s_server2"
	k3sServerThree     = "k3s_server3"
	rke2ServerOne      = "rke2_server1"
	rke2ServerTwo      = "rke2_server2"
	rke2ServerThree    = "rke2_server3"
	rke2InstanceIDs    = "rke2_instance_ids"
	allowUnverifiedSSL = "allow_unverified_ssl"
)

// CreateVsphereTerraformProviderBlock will up the terraform block with the required vsphere provider.
func CreateVsphereTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	vsphereProviderVersion := os.Getenv("VSPHERE_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(defaults.Vsphere, cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.VsphereSource),
		defaults.Version: cty.StringVal(vsphereProviderVersion),
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

	var instanceIds map[string]any
	if terraformConfig.Standalone.RKE2Version != "" {
		instanceIds = map[string]any{
			rke2ServerOne:   defaults.VsphereVirutalMachine + "." + rke2ServerOne + ".id",
			rke2ServerTwo:   defaults.VsphereVirutalMachine + "." + rke2ServerTwo + ".id",
			rke2ServerThree: defaults.VsphereVirutalMachine + "." + rke2ServerThree + ".id",
		}
	} else if terraformConfig.Standalone.K3SVersion != "" {
		instanceIds = map[string]any{
			k3sServerOne:   defaults.VsphereVirutalMachine + "." + k3sServerOne + ".id",
			k3sServerTwo:   defaults.VsphereVirutalMachine + "." + k3sServerTwo + ".id",
			k3sServerThree: defaults.VsphereVirutalMachine + "." + k3sServerThree + ".id",
		}
	}

	instanceIdsBlock := localBlockBody.AppendNewBlock(rke2InstanceIDs+" =", nil)
	instanceIdsBlockBody := instanceIdsBlock.Body()

	for key, value := range instanceIds {
		expression := value.(string)
		instanceValues := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
		}

		instanceIdsBlockBody.SetAttributeRaw(key, instanceValues)
	}
}
