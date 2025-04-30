package rbac

import (
	"os"
	"strings"

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

// RoleCheck is a helper function that will check if the RBAC role is either `clusterOwner` or `projectOwner`.
func RoleCheck(client *rancher.Client, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File, terraform *config.TerraformConfig,
	rbacRole config.Role, isRKE1 bool) (*hclwrite.File, *hclwrite.Body, error) {
	if strings.Contains(string(rbacRole), string(config.ClusterOwner)) {
		newFile, rootBody, err := addClusterRole(client, newFile, rootBody, terraform, rbacRole, isRKE1)
		if err != nil {
			return newFile, rootBody, err
		}
	} else if strings.Contains(string(rbacRole), string(config.ProjectOwner)) {
		newFile, rootBody, err := addProjectMember(client, newFile, rootBody, terraform, rbacRole, isRKE1)
		if err != nil {
			return newFile, rootBody, err
		}
	}

	return newFile, rootBody, nil
}

// addClusterRole is a helper function that will add the RBAC cluster role to non `user` member in the main.tf file.
func addClusterRole(client *rancher.Client, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rbacRole config.Role, isRKE1 bool) (*hclwrite.File, *hclwrite.Body, error) {
	user, err := SetUsers(newFile, rootBody, rbacRole)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	clusterRoleTemplateBindingBlock := rootBody.AppendNewBlock(defaults.Resource, []string{clusterRoleTemplateBinding, terraformConfig.ResourcePrefix})
	clusterRoleTemplateBindingBlockBody := clusterRoleTemplateBindingBlock.Body()

	if isRKE1 {
		clusterBlockID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Cluster + "." + terraformConfig.ResourcePrefix + ".id")},
		}

		clusterRoleTemplateBindingBlockBody.SetAttributeRaw(clusterID, clusterBlockID)
	} else {
		clusterBlockID, err := clusters.GetClusterIDByName(client, terraformConfig.ResourcePrefix)
		if err != nil {
			return newFile, rootBody, err
		}

		logrus.Infof("Cluster ID: %s", clusterBlockID)

		clusterRoleTemplateBindingBlockBody.SetAttributeValue(clusterID, cty.StringVal(clusterBlockID))
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(clusterRoleTemplateBindingName))
	clusterRoleTemplateBindingBlockBody.SetAttributeValue(roleTemplateID, cty.StringVal(string(rbacRole)))

	newUser := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancherUser + "." + user + ".id")},
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeRaw(userID, newUser)

	var dependsOn string
	if isRKE1 {
		dependsOn = `[` + defaults.Cluster + `.` + terraformConfig.ResourcePrefix + `]`
	} else {
		dependsOn = `[` + defaults.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
	}

	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOn)},
	}

	clusterRoleTemplateBindingBlockBody.SetAttributeRaw(defaults.DependsOn, value)

	return newFile, rootBody, nil
}

// addProjectMember is a helper function that will add the RBAC project member to `user` in the main.tf file.
func addProjectMember(client *rancher.Client, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rbacRole config.Role, isRKE1 bool) (*hclwrite.File, *hclwrite.Body, error) {
	user, err := SetUsers(newFile, rootBody, rbacRole)
	if err != nil {
		return nil, nil, err
	}

	rootBody.AppendNewline()

	projectBlock := rootBody.AppendNewBlock(defaults.Resource, []string{project, terraformConfig.ResourcePrefix})
	projectBlockBody := projectBlock.Body()

	projectBlockBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(projectName))

	if isRKE1 {
		clusterBlockID := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(defaults.Cluster + "." + terraformConfig.ResourcePrefix + ".id")},
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
		dependsOn = `[` + defaults.Cluster + `.` + terraformConfig.ResourcePrefix + `]`
	} else {
		dependsOn = `[` + defaults.ClusterV2 + `.` + terraformConfig.ResourcePrefix + `]`
	}

	value := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOn)},
	}

	projectBlockBody.SetAttributeRaw(defaults.DependsOn, value)

	projectRoleTemplateBindingBlock := rootBody.AppendNewBlock(defaults.Resource, []string{projectRoleTemplateBinding, terraformConfig.ResourcePrefix})
	projectRoleTemplateBindingBody := projectRoleTemplateBindingBlock.Body()

	projectRoleTemplateBindingBody.SetAttributeValue(defaults.ResourceName, cty.StringVal(projectRoleTemplateBindingName))

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

	projectRoleTemplateBindingBody.SetAttributeRaw(defaults.DependsOn, value)

	return newFile, rootBody, nil
}
