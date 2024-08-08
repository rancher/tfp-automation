package authproviders

type GithubConfig struct {
	ClientID     string `json:"clientID,omitempty" yaml:"clientID,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
}
