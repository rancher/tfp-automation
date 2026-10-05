package authproviders

type KeycloakOIDCConfig struct {
	ClientID           string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret       string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	Issuer             string `json:"issuer,omitempty" yaml:"issuer,omitempty"`
	RancherURL         string `json:"rancherUrl,omitempty" yaml:"rancherUrl,omitempty"`
	AuthEndpoint       string `json:"authEndpoint,omitempty" yaml:"authEndpoint,omitempty"`
	TokenEndpoint      string `json:"tokenEndpoint,omitempty" yaml:"tokenEndpoint,omitempty"`
	UserInfoEndpoint   string `json:"userInfoEndpoint,omitempty" yaml:"userInfoEndpoint,omitempty"`
	JWKSUrl            string `json:"jwksUrl,omitempty" yaml:"jwksUrl,omitempty"`
	EndSessionEndpoint string `json:"endSessionEndpoint,omitempty" yaml:"endSessionEndpoint,omitempty"`
	Scopes             string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	GroupsField        string `json:"groupsField,omitempty" yaml:"groupsField,omitempty"`
	GroupSearchEnabled *bool  `json:"groupSearchEnabled,omitempty" yaml:"groupSearchEnabled,omitempty"`
	Certificate        string `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	PrivateKey         string `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`

	AccessMode          string   `json:"accessMode,omitempty" yaml:"accessMode,omitempty"`
	AllowedPrincipalIDs []string `json:"allowedPrincipalIds,omitempty" yaml:"allowedPrincipalIds,omitempty"`
	Enabled             *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}
