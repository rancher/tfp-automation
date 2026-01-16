package rbac

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// addClusterRole is a helper function that will add the RBAC cluster role to non `user` member in the main.tf file.
func addClusterRole(client *rancher.Client, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rbacRole config.Role, isRKE1 bool) (*hclwrite.File, *hclwrite.Body, error) {
	user, err := setUsers(newFile, rootBody, rbacRole)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	clusterRoleTemplateBindingBlock := rootBody.AppendNewBlock(general.Resource, []string{clusterRoleTemplateBinding, terraformConfig.ResourcePrefix})
	clusterRoleTemplateBindingBlockBody := clusterRoleTemplateBindingBlock.Body()

	if isRKE1 {
		clusterBlockID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(rancher2.Cluster + "." + terraformConfig.ResourcePrefix + ".id")},
		}

		clusterRoleTemplateBindingBlockBody.SetAttributeRaw(clusterID, clusterBlockID)
	} else {
		clusterBlockID, err := clusters.GetClusterIDByName(client, terraformConfig.ResourcePrefix)
		if err != nil {
			return newFile, rootBody, err
		}

		clusterRoleTemplateBindingBlockBody.SetAttributeValue(clusterID, cty.StringVal(clusterBlockID))
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(clusterRoleTemplateBindingName))
	clusterRoleTemplateBindingBlockBody.SetAttributeValue(roleTemplateID, cty.StringVal(string(rbacRole)))

	newUser := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancherUser + "." + user + ".id")},
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeRaw(userID, newUser)

	var dependsOn string
	if isRKE1 {
		dependsOn = `[` + rancher2.Cluster + `.` + terraformConfig.ResourcePrefix + `]`
	} else {
		dependsOn = `[` + rancher2.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
	}

	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOn)},
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeRaw(general.DependsOn, value)

	return newFile, rootBody, nil
}
