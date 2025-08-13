package rbac

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
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
