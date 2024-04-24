package provisioning

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	password "github.com/rancher/shepherd/extensions/users/passwordgenerator"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	blocks "github.com/rancher/tfp-automation/defaults/resourceblocks/providersUsers"
	"github.com/zclconf/go-cty/cty"
)

// SetProvidersAndUsersTF is a helper function that will set the general Terraform configurations in the main.tf file.
func SetProvidersAndUsersTF(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig) (*hclwrite.File, *hclwrite.Body) {
	providerVersion := os.Getenv("RANCHER2_PROVIDER_VERSION")

	source := "rancher/rancher2"
	if strings.Contains(providerVersion, blocks.RC) {
		source = "terraform.local/local/rancher2"
	}

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(blocks.Terraform, nil)
	tfBlockBody := tfBlock.Body()

	reqProvsBlock := tfBlockBody.AppendNewBlock(blocks.RequiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue(blocks.Rancher, cty.ObjectVal(map[string]cty.Value{
		blocks.Source:  cty.StringVal(source),
		blocks.Version: cty.StringVal(providerVersion),
	}))

	rootBody.AppendNewline()

	provBlock := rootBody.AppendNewBlock(blocks.Provider, []string{blocks.Rancher})
	provBlockBody := provBlock.Body()

	provBlockBody.SetAttributeValue(blocks.ApiURL, cty.StringVal(`https://`+rancherConfig.Host))
	provBlockBody.SetAttributeValue(blocks.TokenKey, cty.StringVal(rancherConfig.AdminToken))
	provBlockBody.SetAttributeValue(blocks.Insecure, cty.BoolVal(*rancherConfig.Insecure))

	rootBody.AppendNewline()

	var testuser = namegen.AppendRandomString("testuser")
	var testpassword = password.GenerateUserPassword("testpass")

	userBlock := rootBody.AppendNewBlock(blocks.Resource, []string{blocks.RancherUser, blocks.RancherUser})
	userBlockBody := userBlock.Body()

	userBlockBody.SetAttributeValue(blocks.Name, cty.StringVal(testuser))
	userBlockBody.SetAttributeValue(blocks.Username, cty.StringVal(testuser))
	userBlockBody.SetAttributeValue(blocks.Password, cty.StringVal(testpassword))
	userBlockBody.SetAttributeValue(blocks.Enabled, cty.BoolVal(true))

	rootBody.AppendNewline()

	globalRoleBindingBlock := rootBody.AppendNewBlock(blocks.Resource, []string{blocks.GlobalRoleBinding, blocks.GlobalRoleBinding})
	globalRoleBindingBlockBody := globalRoleBindingBlock.Body()

	globalRoleBindingBlockBody.SetAttributeValue(blocks.Name, cty.StringVal(testuser))
	globalRoleBindingBlockBody.SetAttributeValue(blocks.GlobalRoleID, cty.StringVal(blocks.User))
	globalRoleBindingBlockBody.SetAttributeValue(blocks.UserID, cty.StringVal(blocks.Rancher+blocks.Rancher+".id"))

	return newFile, rootBody
}
