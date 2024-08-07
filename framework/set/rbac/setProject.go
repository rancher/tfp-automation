package rbac

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	project = "rancher2_project"
	cluster = "rancher2_cluster_v2"

	clusterRoleTemplateBinding = "rancher2_cluster_role_template_binding"
	projectRoleTemplateBinding = "rancher2_project_role_template_binding"

	clusterRoleTemplateBindingName = "tfp-cluster-role-template-binding"
	projectName                    = "tfp-project"
	projectRoleTemplateBindingName = "tfp-project-role-template-binding"
	clusterID                      = "cluster_id"
	projectID                      = "project_id"
	roleTemplateID                 = "role_template_id"
)

// AddProjectMember is a helper function that will add the RBAC project member to `user` in the main.tf file.
func AddProjectMember(client *rancher.Client, clusterName string, newFile *hclwrite.File, rootBody *hclwrite.Body,
	clusterBlockID hclwrite.Tokens, rbacRole config.Role, newUser string, isRKE1 bool) (*hclwrite.File, *hclwrite.Body) {
	projectBlock := rootBody.AppendNewBlock(defaults.Resource, []string{project, project})
	projectBlockBody := projectBlock.Body()

	projectBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(projectName))

	if isRKE1 {
		projectBlockBody.SetAttributeRaw(clusterID, clusterBlockID)
	} else {
		clusterBlockID, err := clusters.GetClusterIDByName(client, clusterName)
		if err != nil {
			return newFile, rootBody
		}

		projectBlockBody.SetAttributeValue(clusterID, cty.StringVal(clusterBlockID))
	}

	rootBody.AppendNewline()

	projectRoleTemplateBindingBlock := rootBody.AppendNewBlock(defaults.Resource, []string{projectRoleTemplateBinding, projectRoleTemplateBinding})
	projectRoleTemplateBindingBody := projectRoleTemplateBindingBlock.Body()

	projectBlockID := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(project + "." + project + ".id")},
	}

	projectRoleTemplateBindingBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(projectRoleTemplateBindingName))
	projectRoleTemplateBindingBody.SetAttributeRaw(projectID, projectBlockID)
	projectRoleTemplateBindingBody.SetAttributeValue(roleTemplateID, cty.StringVal(string(rbacRole)))
	projectRoleTemplateBindingBody.SetAttributeValue(userID, cty.StringVal(newUser))

	logrus.Infof("Added project member: %s", projectName)

	return newFile, rootBody
}

// AddClusterRole is a helper function that will add the RBAC cluster role to non `user` member in the main.tf file.
func AddClusterRole(client *rancher.Client, clusterName string, newFile *hclwrite.File, rootBody *hclwrite.Body, clusterBlockID hclwrite.Tokens,
	rbacRole config.Role, newUser string, isRKE1 bool) (*hclwrite.File, *hclwrite.Body) {
	clusterRoleTemplateBindingBlock := rootBody.AppendNewBlock(defaults.Resource, []string{clusterRoleTemplateBinding, clusterRoleTemplateBinding})
	clusterRoleTemplateBindingBlockBody := clusterRoleTemplateBindingBlock.Body()

	if isRKE1 {
		clusterRoleTemplateBindingBlockBody.SetAttributeRaw(clusterID, clusterBlockID)
	} else {
		clusterBlockID, err := clusters.GetClusterIDByName(client, clusterName)
		if err != nil {
			return newFile, rootBody
		}

		clusterRoleTemplateBindingBlockBody.SetAttributeValue(clusterID, cty.StringVal(clusterBlockID))
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterRoleTemplateBindingName))
	clusterRoleTemplateBindingBlockBody.SetAttributeValue(roleTemplateID, cty.StringVal(string(rbacRole)))
	clusterRoleTemplateBindingBlockBody.SetAttributeValue(userID, cty.StringVal(newUser))

	logrus.Infof("Added cluster role: %s", clusterRoleTemplateBindingName)

	return newFile, rootBody
}
