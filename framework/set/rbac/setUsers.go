package rbac

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	password "github.com/rancher/shepherd/extensions/users/passwordgenerator"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

const (
	apiURL            = "api_url"
	globalRoleBinding = "rancher2_global_role_binding"
	globalRoleID      = "global_role_id"
	insecure          = "insecure"
	name              = "name"
	provider          = "provider"
	rancher2Const     = "rancher2"
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

// SetUsers is a helper function that will set the RBAC users in the main.tf file.
func setUsers(newFile *hclwrite.File, rootBody *hclwrite.Body, rbacRole config.Role) (string, error) {
	var testuser = namegen.AppendRandomString("testuser")
	var testpassword = password.GenerateUserPassword("testpass")

	userBlock := rootBody.AppendNewBlock(general.Resource, []string{rancherUser, testuser})
	userBlockBody := userBlock.Body()

	userBlockBody.SetAttributeValue(name, cty.StringVal(testuser))
	userBlockBody.SetAttributeValue(username, cty.StringVal(testuser))
	userBlockBody.SetAttributeValue(testPassword, cty.StringVal(testpassword))
	userBlockBody.SetAttributeValue(clusters.Enabled, cty.BoolVal(true))

	rootBody.AppendNewline()

	globalRoleBindingBlock := rootBody.AppendNewBlock(general.Resource, []string{globalRoleBinding, testuser})
	globalRoleBindingBlockBody := globalRoleBindingBlock.Body()

	globalRoleBindingBlockBody.SetAttributeValue(name, cty.StringVal(testuser))
	globalRoleBindingBlockBody.SetAttributeValue(globalRoleID, cty.StringVal(user))

	user := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancherUser + "." + testuser + ".id")},
	}

	globalRoleBindingBlockBody.SetAttributeRaw(userID, user)

	return testuser, nil
}
