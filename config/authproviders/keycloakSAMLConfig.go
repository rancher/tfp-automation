package authproviders

type KeycloakSAMLConfig struct {
	DisplayNameField   string `json:"displayNameField,omitempty" yaml:"displayNameField,omitempty"`
	GroupsField        string `json:"groupsField,omitempty" yaml:"groupsField,omitempty"`
	IdpMetadataContent string `json:"idpMetadataContent,omitempty" yaml:"idpMetadataContent,omitempty"`
	RancherAPIHost     string `json:"rancherApiHost,omitempty" yaml:"rancherApiHost,omitempty"`
	EntityID           string `json:"entityID,omitempty" yaml:"entityID,omitempty"`
	SPCert             string `json:"spCert,omitempty" yaml:"spCert,omitempty"`
	SPKey              string `json:"spKey,omitempty" yaml:"spKey,omitempty"`
	UIDField           string `json:"uidField,omitempty" yaml:"uidField,omitempty"`
	UserNameField      string `json:"userNameField,omitempty" yaml:"userNameField,omitempty"`

	AccessMode          string   `json:"accessMode,omitempty" yaml:"accessMode,omitempty"`
	AllowedPrincipalIDs []string `json:"allowedPrincipalIds,omitempty" yaml:"allowedPrincipalIds,omitempty"`
	Enabled             *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}
