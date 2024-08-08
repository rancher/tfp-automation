package ad

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	adConfig = "rancher2_auth_config_activedirectory"

	resource               = "resource"
	port                   = "port"
	servers                = "servers"
	serviceAccountPassword = "service_account_password"
	serviceAccountUsername = "service_account_username"
	userSearchBase         = "user_search_base"
	testUsername           = "test_username"
	testPassword           = "test_password"
)

// SetAD is a function that will set the AD configurations in the main.tf file.
func SetAD(terraformConfig *config.TerraformConfig, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) error {
	adBlock := rootBody.AppendNewBlock(resource, []string{adConfig, adConfig})
	adBlockBody := adBlock.Body()

	adBlockBody.SetAttributeValue(port, cty.NumberIntVal(int64(terraformConfig.ADConfig.Port)))
	adBlockBody.SetAttributeValue(servers, cty.ListVal([]cty.Value{cty.StringVal(terraformConfig.ADConfig.Servers[0])}))
	adBlockBody.SetAttributeValue(serviceAccountPassword, cty.StringVal(terraformConfig.ADConfig.ServiceAccountPassword))
	adBlockBody.SetAttributeValue(serviceAccountUsername, cty.StringVal(terraformConfig.ADConfig.ServiceAccountUsername))
	adBlockBody.SetAttributeValue(userSearchBase, cty.StringVal(terraformConfig.ADConfig.UserSearchBase))
	adBlockBody.SetAttributeValue(testUsername, cty.StringVal(terraformConfig.ADConfig.TestUsername))
	adBlockBody.SetAttributeValue(testPassword, cty.StringVal(terraformConfig.ADConfig.TestPassword))

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write Active Directory configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
