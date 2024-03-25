package nodeproviders

type GoogleConfig struct {
	AuthEncodedJSON string `json:"authEncodedJson" yaml:"authEncodedJson"`
	Network         string `json:"network,omitempty" yaml:"network,omitempty"`
	ProjectID       string `json:"projectID,omitempty" yaml:"projectID,omitempty"`
	Subnetwork      string `json:"subnetwork,omitempty" yaml:"subnetwork,omitempty"`
	Region          string `json:"region,omitempty" yaml:"region,omitempty"`
}
