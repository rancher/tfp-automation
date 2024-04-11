package provisioning

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	password "github.com/rancher/shepherd/extensions/users/passwordgenerator"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/zclconf/go-cty/cty"
)

// SetProvidersAndUsersTF is a helper function that will set the general Terraform configurations in the main.tf file.
func SetProvidersAndUsersTF(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig) (*hclwrite.File, *hclwrite.Body) {
	providerVersion := os.Getenv("RANCHER2_PROVIDER_VERSION")

	source := "rancher/rancher2"
	if strings.Contains(providerVersion, "-rc") {
		source = "terraform.local/local/rancher2"
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

	rootBody.AppendNewline()

	var testuser = namegen.AppendRandomString("testuser")
	var testpassword = password.GenerateUserPassword("testpass")

	userBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_user", "rancher2_user"})
	userBlockBody := userBlock.Body()

	userBlockBody.SetAttributeValue("name", cty.StringVal(testuser))
	userBlockBody.SetAttributeValue("username", cty.StringVal(testuser))
	userBlockBody.SetAttributeValue("password", cty.StringVal(testpassword))
	userBlockBody.SetAttributeValue("enabled", cty.BoolVal(true))

	rootBody.AppendNewline()

	globalRoleBindingBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_global_role_binding", "rancher2_global_role_binding"})
	globalRoleBindingBlockBody := globalRoleBindingBlock.Body()

	globalRoleBindingBlockBody.SetAttributeValue("name", cty.StringVal(testuser))
	globalRoleBindingBlockBody.SetAttributeValue("global_role_id", cty.StringVal("user"))
	globalRoleBindingBlockBody.SetAttributeValue("user_id", cty.StringVal("rancher2_user.rancher2_user.id"))

	return newFile, rootBody
}
