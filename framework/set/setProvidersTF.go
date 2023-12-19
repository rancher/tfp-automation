package framework

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// setProvidersTF is a helper function that will set the general Terraform configurations in the main.tf file.
func setProvidersTF(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig) (*hclwrite.File, *hclwrite.Body) {
	source := "rancher/rancher2"
	providerVersion := os.Getenv("RANCHER2_PROVIDER_VERSION")
	
	if strings.Contains(providerVersion, "-rc") { 
		// source = "terraform.local/local/rancher2"
		source = "/home/jenkins/workspace/rancher_qa/tfp-automation/.terraform.d/plugins/terraform.local/local/rancher2"
	} 
	
	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock("terraform", nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock("required_providers", nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue("rancher2", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal(source),
		"version": cty.StringVal(providerVersion),
	}))

	rootBody.AppendNewline()

	provBlock := rootBody.AppendNewBlock("provider", []string{"rancher2"})
	provBlockBody := provBlock.Body()

	provBlockBody.SetAttributeValue("api_url", cty.StringVal(`https://`+rancherConfig.Host))
	provBlockBody.SetAttributeValue("token_key", cty.StringVal(rancherConfig.AdminToken))
	provBlockBody.SetAttributeValue("insecure", cty.BoolVal(*rancherConfig.Insecure))

	return newFile, rootBody
}
