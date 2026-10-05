package keycloakOIDC

import (
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	keycloakOIDCConfig = "rancher2_auth_config_keycloak_oidc"

	resource            = "resource"
	clientID            = "client_id"
	clientSecret        = "client_secret"
	issuer              = "issuer"
	rancherURL          = "rancher_url"
	authEndpoint        = "auth_endpoint"
	tokenEndpoint       = "token_endpoint"
	userInfoEndpoint    = "userinfo_endpoint"
	jwksURL             = "jwks_url"
	endSessionEndpoint  = "end_session_endpoint"
	scopes              = "scopes"
	groupsField         = "groups_field"
	groupSearchEnabled  = "group_search_enabled"
	certificate         = "certificate"
	privateKey          = "private_key"
	accessMode          = "access_mode"
	allowedPrincipalIDs = "allowed_principal_ids"
	enabled             = "enabled"

	httpsPrefix    = "https://"
	verifyAuthPath = "/verify-auth"
)

// SetKeycloakOIDC is a function that will set the Keycloak OIDC configurations in the main.tf file.
func SetKeycloakOIDC(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) error {
	keycloakBlock := rootBody.AppendNewBlock(resource, []string{keycloakOIDCConfig, keycloakOIDCConfig})
	keycloakBlockBody := keycloakBlock.Body()

	url := terraformConfig.KeycloakOIDCConfig.RancherURL
	if url == "" {
		url = httpsPrefix + rancherConfig.Host + verifyAuthPath
	}

	keycloakBlockBody.SetAttributeValue(clientID, cty.StringVal(terraformConfig.KeycloakOIDCConfig.ClientID))
	keycloakBlockBody.SetAttributeValue(clientSecret, cty.StringVal(terraformConfig.KeycloakOIDCConfig.ClientSecret))
	keycloakBlockBody.SetAttributeValue(issuer, cty.StringVal(terraformConfig.KeycloakOIDCConfig.Issuer))
	keycloakBlockBody.SetAttributeValue(rancherURL, cty.StringVal(url))

	optional := map[string]string{
		authEndpoint:       terraformConfig.KeycloakOIDCConfig.AuthEndpoint,
		tokenEndpoint:      terraformConfig.KeycloakOIDCConfig.TokenEndpoint,
		userInfoEndpoint:   terraformConfig.KeycloakOIDCConfig.UserInfoEndpoint,
		jwksURL:            terraformConfig.KeycloakOIDCConfig.JWKSUrl,
		endSessionEndpoint: terraformConfig.KeycloakOIDCConfig.EndSessionEndpoint,
		scopes:             terraformConfig.KeycloakOIDCConfig.Scopes,
		groupsField:        terraformConfig.KeycloakOIDCConfig.GroupsField,
		certificate:        terraformConfig.KeycloakOIDCConfig.Certificate,
		privateKey:         terraformConfig.KeycloakOIDCConfig.PrivateKey,
		accessMode:         terraformConfig.KeycloakOIDCConfig.AccessMode,
	}

	for attribute, value := range optional {
		if value != "" {
			keycloakBlockBody.SetAttributeValue(attribute, cty.StringVal(value))
		}
	}

	if terraformConfig.KeycloakOIDCConfig.GroupSearchEnabled != nil {
		keycloakBlockBody.SetAttributeValue(groupSearchEnabled, cty.BoolVal(*terraformConfig.KeycloakOIDCConfig.GroupSearchEnabled))
	}

	if len(terraformConfig.KeycloakOIDCConfig.AllowedPrincipalIDs) > 0 {
		principalIDs := []cty.Value{}
		for _, principalID := range terraformConfig.KeycloakOIDCConfig.AllowedPrincipalIDs {
			principalIDs = append(principalIDs, cty.StringVal(principalID))
		}

		keycloakBlockBody.SetAttributeValue(allowedPrincipalIDs, cty.ListVal(principalIDs))
	}

	if terraformConfig.KeycloakOIDCConfig.Enabled != nil {
		keycloakBlockBody.SetAttributeValue(enabled, cty.BoolVal(*terraformConfig.KeycloakOIDCConfig.Enabled))
	}

	_, err := file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write Keycloak OIDC configurations to main.tf file. Error: %v", err)
		return err
	}

	return nil
}
