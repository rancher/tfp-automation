package authproviders

type AzureADConfig struct {
	ApplicationID     string `json:"applicationID,omitempty" yaml:"applicationID,omitempty"`
	ApplicationSecret string `json:"applicationSecret" yaml:"applicationSecret,omitempty"`
	AuthEndpoint      string `json:"authEndpoint,omitempty" yaml:"authEndpoint,omitempty"`
	GraphEndpoint     string `json:"graphEndpoint,omitempty" yaml:"graphEndpoint,omitempty"`
	TenantID          string `json:"tenantID,omitempty" yaml:"tenantID,omitempty"`
	TokenEndpoint     string `json:"tokenEndpoint" yaml:"tokenEndpoint,omitempty"`
}
