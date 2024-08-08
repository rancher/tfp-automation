package authproviders

type ADConfig struct {
	Port                   int64    `json:"port,omitempty" yaml:"port,omitempty"`
	Servers                []string `json:"servers,omitempty" yaml:"servers,omitempty"`
	ServiceAccountPassword string   `json:"serviceAccountPassword,omitempty" yaml:"serviceAccountPassword,omitempty"`
	ServiceAccountUsername string   `json:"serviceAccountUsername,omitempty" yaml:"serviceAccountUsername,omitempty"`
	UserSearchBase         string   `json:"userSearchBase,omitempty" yaml:"userSearchBase,omitempty"`
	TestUsername           string   `json:"testUsername,omitempty" yaml:"testUsername,omitempty"`
	TestPassword           string   `json:"testPassword,omitempty" yaml:"testPassword,omitempty"`
}
