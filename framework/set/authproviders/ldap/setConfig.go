package ldap

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	openLDAPConfig = "rancher2_auth_config_openldap"

	resource                       = "resource"
	port                           = "port"
	servers                        = "servers"
	serviceAccountDistinguisedName = "service_account_distinguished_name"
	serviceAccountPassword         = "service_account_password"
	userSearchBase                 = "user_search_base"
	testUsername                   = "test_username"
	testPassword                   = "test_password"
)

// SetOpenLDAP is a function that will set the OpenLDAP configurations in the main.tf file.
func SetOpenLDAP(terraformConfig *config.TerraformConfig, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) error {
	openLDAPBlock := rootBody.AppendNewBlock(resource, []string{openLDAPConfig, openLDAPConfig})
	openLDAPBlockBody := openLDAPBlock.Body()

	openLDAPBlockBody.SetAttributeValue(port, cty.NumberIntVal(int64(terraformConfig.OpenLDAPConfig.Port)))
	openLDAPBlockBody.SetAttributeValue(servers, cty.ListVal([]cty.Value{cty.StringVal(terraformConfig.OpenLDAPConfig.Servers[0])}))
	openLDAPBlockBody.SetAttributeValue(serviceAccountDistinguisedName, cty.StringVal(terraformConfig.OpenLDAPConfig.ServiceAccountDistinguisedName))
	openLDAPBlockBody.SetAttributeValue(serviceAccountPassword, cty.StringVal(terraformConfig.OpenLDAPConfig.ServiceAccountPassword))
	openLDAPBlockBody.SetAttributeValue(userSearchBase, cty.StringVal(terraformConfig.OpenLDAPConfig.UserSearchBase))
	openLDAPBlockBody.SetAttributeValue(testUsername, cty.StringVal(terraformConfig.OpenLDAPConfig.TestUsername))
	openLDAPBlockBody.SetAttributeValue(testPassword, cty.StringVal(terraformConfig.OpenLDAPConfig.TestPassword))

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write OpenLDAP configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
