package authproviders

type ADConfig struct {
	Port                   int64    `json:"port,omitempty" yaml:"port,omitempty"`
	Servers                []string `json:"servers,omitempty" yaml:"servers,omitempty"`
	ServiceAccountPassword string   `json:"serviceAccountPassword,omitempty" yaml:"serviceAccountPassword,omitempty"`
	ServiceAccountUsername string   `json:"serviceAccountUsername,omitempty" yaml:"serviceAccountUsername,omitempty"`
	UserSearchBase         string   `json:"userSearchBase,omitempty" yaml:"userSearchBase,omitempty"`
	TestUsername           string   `json:"testUsername,omitempty" yaml:"testUsername,omitempty"`
	TestPassword           string   `json:"testPassword,omitempty" yaml:"testPassword,omitempty"`

	GroupSearchBase              string `json:"groupSearchBase,omitempty" yaml:"groupSearchBase,omitempty"`
	Enabled                      *bool  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	TLS                          *bool  `json:"tls,omitempty" yaml:"tls,omitempty"`
	StartTLS                     *bool  `json:"startTLS,omitempty" yaml:"startTLS,omitempty"`
	NestedGroupMembershipEnabled *bool  `json:"nestedGroupMembershipEnabled,omitempty" yaml:"nestedGroupMembershipEnabled,omitempty"`
}
