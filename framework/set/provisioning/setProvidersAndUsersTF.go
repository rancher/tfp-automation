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

const (
	apiURL            = "api_url"
	globalRoleBinding = "rancher2_global_role_binding"
	globalRoleID      = "global_role_id"
	insecure          = "insecure"
	name              = "name"
	provider          = "provider"
	rancher2          = "rancher2"
	rancherSource     = "source"
	rancherUser       = "rancher2_user"
	rc                = "-rc"
	requiredProviders = "required_providers"
	terraform         = "terraform"
	testPassword      = "password"
	tokenKey          = "token_key"
	version           = "version"
	user              = "user"
	userID            = "user_id"
	username          = "username"
)

// SetProvidersAndUsersTF is a helper function that will set the general Terraform configurations in the main.tf file.
func SetProvidersAndUsersTF(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig) (*hclwrite.File, *hclwrite.Body) {
	providerVersion := os.Getenv("RANCHER2_PROVIDER_VERSION")

	source := "rancher/rancher2"
	if strings.Contains(providerVersion, rc) {
		source = "terraform.local/local/rancher2"
	}

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraform, nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(rancher2, cty.ObjectVal(map[string]cty.Value{
		rancherSource: cty.StringVal(source),
		version:       cty.StringVal(providerVersion),
	}))

	rootBody.AppendNewline()

	provBlock := rootBody.AppendNewBlock(provider, []string{rancher2})
	provBlockBody := provBlock.Body()

	provBlockBody.SetAttributeValue(apiURL, cty.StringVal(`https://`+rancherConfig.Host))
	provBlockBody.SetAttributeValue(tokenKey, cty.StringVal(rancherConfig.AdminToken))
	provBlockBody.SetAttributeValue(insecure, cty.BoolVal(*rancherConfig.Insecure))

	rootBody.AppendNewline()

	var testuser = namegen.AppendRandomString("testuser")
	var testpassword = password.GenerateUserPassword("testpass")

	userBlock := rootBody.AppendNewBlock(resource, []string{rancherUser, rancherUser})
	userBlockBody := userBlock.Body()

	userBlockBody.SetAttributeValue(name, cty.StringVal(testuser))
	userBlockBody.SetAttributeValue(username, cty.StringVal(testuser))
	userBlockBody.SetAttributeValue(testPassword, cty.StringVal(testpassword))
	userBlockBody.SetAttributeValue(enabled, cty.BoolVal(true))

	rootBody.AppendNewline()

	globalRoleBindingBlock := rootBody.AppendNewBlock(resource, []string{globalRoleBinding, globalRoleBinding})
	globalRoleBindingBlockBody := globalRoleBindingBlock.Body()

	globalRoleBindingBlockBody.SetAttributeValue(name, cty.StringVal(testuser))
	globalRoleBindingBlockBody.SetAttributeValue(globalRoleID, cty.StringVal(user))
	globalRoleBindingBlockBody.SetAttributeValue(userID, cty.StringVal(rancher2+rancher2+".id"))

	return newFile, rootBody
}
