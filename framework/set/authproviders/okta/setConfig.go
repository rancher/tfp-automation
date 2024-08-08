package okta

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	oktaConfig = "rancher2_auth_config_okta"

	resource           = "resource"
	displayNameField   = "display_name_field"
	groupsField        = "groups_field"
	idpMetadataContent = "idp_metadata_content"
	rancherAPIHost     = "rancher_api_host"
	spCert             = "sp_cert"
	spKey              = "sp_key"
	uidField           = "uid_field"
	userNameField      = "user_name_field"
)

// SetOkta is a function that will set the Okta configurations in the main.tf file.
func SetOkta(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) error {
	oktaBlock := rootBody.AppendNewBlock(resource, []string{oktaConfig, oktaConfig})
	oktaBlockBody := oktaBlock.Body()

	oktaBlockBody.SetAttributeValue(displayNameField, cty.StringVal(terraformConfig.OktaConfig.DisplayNameField))
	oktaBlockBody.SetAttributeValue(groupsField, cty.StringVal(terraformConfig.OktaConfig.GroupsField))
	oktaBlockBody.SetAttributeValue(idpMetadataContent, cty.StringVal(terraformConfig.OktaConfig.IdpMetadataContent))
	oktaBlockBody.SetAttributeValue(rancherAPIHost, cty.StringVal("https://"+rancherConfig.Host))
	oktaBlockBody.SetAttributeValue(spCert, cty.StringVal(terraformConfig.OktaConfig.SPCert))
	oktaBlockBody.SetAttributeValue(spKey, cty.StringVal(terraformConfig.OktaConfig.SPKey))
	oktaBlockBody.SetAttributeValue(uidField, cty.StringVal(terraformConfig.OktaConfig.UIDField))
	oktaBlockBody.SetAttributeValue(userNameField, cty.StringVal(terraformConfig.OktaConfig.UserNameField))

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write Okta configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
