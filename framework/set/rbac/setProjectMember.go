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

// addProjectMember is a helper function that will add the RBAC project member to `user` in the main.tf file.
func addProjectMember(client *rancher.Client, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rbacRole config.Role, isRKE1 bool) (*hclwrite.File, *hclwrite.Body, error) {
	user, err := setUsers(newFile, rootBody, rbacRole)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	projectBlock := rootBody.AppendNewBlock(general.Resource, []string{project, terraformConfig.ResourcePrefix})
	projectBlockBody := projectBlock.Body()

	projectBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(projectName))

	if isRKE1 {
		clusterBlockID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(rancher2.Cluster + "." + terraformConfig.ResourcePrefix + ".id")},
		}

		projectBlockBody.SetAttributeRaw(clusterID, clusterBlockID)
	} else {
		clusterBlockID, err := clusters.GetClusterIDByName(client, terraformConfig.ResourcePrefix)
		if err != nil {
			return newFile, rootBody, err
		}

		projectBlockBody.SetAttributeValue(clusterID, cty.StringVal(clusterBlockID))
	}

	rootBody.AppendNewline()

	var dependsOn string
	if isRKE1 {
		dependsOn = `[` + rancher2.Cluster + `.` + terraformConfig.ResourcePrefix + `]`
	} else {
		dependsOn = `[` + rancher2.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
	}

	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOn)},
	}

	projectBlockBody.SetAttributeRaw(general.DependsOn, value)

	projectRoleTemplateBindingBlock := rootBody.AppendNewBlock(general.Resource, []string{projectRoleTemplateBinding, terraformConfig.ResourcePrefix})
	projectRoleTemplateBindingBody := projectRoleTemplateBindingBlock.Body()

	projectRoleTemplateBindingBody.SetAttributeValue(general.ResourceName, cty.StringVal(projectRoleTemplateBindingName))

	projectBlockID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(project + "." + terraformConfig.ResourcePrefix + ".id")},
	}

	projectRoleTemplateBindingBody.SetAttributeRaw(projectID, projectBlockID)
	projectRoleTemplateBindingBody.SetAttributeValue(roleTemplateID, cty.StringVal(string(rbacRole)))

	newUser := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancherUser + "." + user + ".id")},
	}

	projectRoleTemplateBindingBody.SetAttributeRaw(userID, newUser)

	if isRKE1 {
		dependsOn = `[` + projectRoleTemplateBinding + `.` + terraformConfig.ResourcePrefix + `]`
	} else {
		dependsOn = `[` + projectRoleTemplateBinding + `.` + terraformConfig.ResourcePrefix + `]`
	}

	value = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOn)},
	}

	projectRoleTemplateBindingBody.SetAttributeRaw(general.DependsOn, value)

	return newFile, rootBody, nil
}
