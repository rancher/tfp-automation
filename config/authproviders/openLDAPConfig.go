package authproviders

type OpenLDAPConfig struct {
	Port                           int64    `json:"port,omitempty" yaml:"port,omitempty"`
	Servers                        []string `json:"servers,omitempty" yaml:"servers,omitempty"`
	ServiceAccountDistinguisedName string   `json:"serviceAccountDistinguishedName,omitempty" yaml:"serviceAccountDistinguishedName,omitempty"`
	ServiceAccountPassword         string   `json:"serviceAccountPassword,omitempty" yaml:"serviceAccountPassword,omitempty"`
	UserSearchBase                 string   `json:"userSearchBase,omitempty" yaml:"userSearchBase,omitempty"`
	TestUsername                   string   `json:"testUsername,omitempty" yaml:"testUsername,omitempty"`
	TestPassword                   string   `json:"testPassword,omitempty" yaml:"testPassword,omitempty"`
}
