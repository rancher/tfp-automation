package keycloakSAML

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	keycloakConfig = "rancher2_auth_config_keycloak"

	resource            = "resource"
	displayNameField    = "display_name_field"
	groupsField         = "groups_field"
	idpMetadataContent  = "idp_metadata_content"
	rancherAPIHost      = "rancher_api_host"
	spCert              = "sp_cert"
	spKey               = "sp_key"
	uidField            = "uid_field"
	userNameField       = "user_name_field"
	entityID            = "entity_id"
	accessMode          = "access_mode"
	allowedPrincipalIDs = "allowed_principal_ids"
	enabled             = "enabled"

	httpsPrefix = "https://"
)

// SetKeycloakSAML is a function that will set the Keycloak SAML configurations in the main.tf file.
func SetKeycloakSAML(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) error {
	keycloakBlock := rootBody.AppendNewBlock(resource, []string{keycloakConfig, keycloakConfig})
	keycloakBlockBody := keycloakBlock.Body()

	host := terraformConfig.KeycloakSAMLConfig.RancherAPIHost
	if host == "" {
		host = httpsPrefix + rancherConfig.Host
	}

	keycloakBlockBody.SetAttributeValue(displayNameField, cty.StringVal(terraformConfig.KeycloakSAMLConfig.DisplayNameField))
	keycloakBlockBody.SetAttributeValue(groupsField, cty.StringVal(terraformConfig.KeycloakSAMLConfig.GroupsField))
	keycloakBlockBody.SetAttributeValue(idpMetadataContent, cty.StringVal(terraformConfig.KeycloakSAMLConfig.IdpMetadataContent))
	keycloakBlockBody.SetAttributeValue(rancherAPIHost, cty.StringVal(host))
	keycloakBlockBody.SetAttributeValue(spCert, cty.StringVal(terraformConfig.KeycloakSAMLConfig.SPCert))
	keycloakBlockBody.SetAttributeValue(spKey, cty.StringVal(terraformConfig.KeycloakSAMLConfig.SPKey))
	keycloakBlockBody.SetAttributeValue(uidField, cty.StringVal(terraformConfig.KeycloakSAMLConfig.UIDField))
	keycloakBlockBody.SetAttributeValue(userNameField, cty.StringVal(terraformConfig.KeycloakSAMLConfig.UserNameField))

	if terraformConfig.KeycloakSAMLConfig.EntityID != "" {
		keycloakBlockBody.SetAttributeValue(entityID, cty.StringVal(terraformConfig.KeycloakSAMLConfig.EntityID))
	}

	if terraformConfig.KeycloakSAMLConfig.AccessMode != "" {
		keycloakBlockBody.SetAttributeValue(accessMode, cty.StringVal(terraformConfig.KeycloakSAMLConfig.AccessMode))
	}

	if len(terraformConfig.KeycloakSAMLConfig.AllowedPrincipalIDs) > 0 {
		principalIDs := []cty.Value{}
		for _, principalID := range terraformConfig.KeycloakSAMLConfig.AllowedPrincipalIDs {
			principalIDs = append(principalIDs, cty.StringVal(principalID))
		}

		keycloakBlockBody.SetAttributeValue(allowedPrincipalIDs, cty.ListVal(principalIDs))
	}

	if terraformConfig.KeycloakSAMLConfig.Enabled != nil {
		keycloakBlockBody.SetAttributeValue(enabled, cty.BoolVal(*terraformConfig.KeycloakSAMLConfig.Enabled))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write Keycloak SAML configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
