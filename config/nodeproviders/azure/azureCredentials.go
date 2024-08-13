package azure

type Credentials struct {
	ClientID       string `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret   string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	Environment    string `json:"environment,omitempty" yaml:"environment,omitempty"`
	SubscriptionID string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	TenantID       string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
}
