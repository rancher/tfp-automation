package authproviders

type OktaConfig struct {
	DisplayNameField   string `json:"displayNameField,omitempty" yaml:"displayNameField,omitempty"`
	GroupsField        string `json:"groupsField,omitempty" yaml:"groupsField,omitempty"`
	IdpMetadataContent string `json:"idpMetadataContent,omitempty" yaml:"idpMetadataContent,omitempty"`
	SPCert             string `json:"spCert,omitempty" yaml:"spCert,omitempty"`
	SPKey              string `json:"spKey,omitempty" yaml:"spKey,omitempty"`
	UIDField           string `json:"uidField,omitempty" yaml:"uidField,omitempty"`
	UserNameField      string `json:"userNameField,omitempty" yaml:"userNameField,omitempty"`
}
