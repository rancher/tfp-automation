package nodeproviders

type GoogleConfig struct {
	GKENetwork    string `json:"gkeNetwork,omitempty" yaml:"gkeNetwork,omitempty"`
	GKEProjectID  string `json:"gkeProjectID,omitempty" yaml:"gkeProjectID,omitempty"`
	GKESubnetwork string `json:"gkeSubnetwork,omitempty" yaml:"gkeSubnetwork,omitempty"`
	Region        string `json:"region,omitempty" yaml:"region,omitempty"`
}
