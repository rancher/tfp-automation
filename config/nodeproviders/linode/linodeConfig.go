package linode

type Config struct {
	LinodeImage    string `json:"linodeImage,omitempty" yaml:"linodeImage,omitempty"`
	LinodeRootPass string `json:"linodeRootPass,omitempty" yaml:"linodeRootPass,omitempty"`
	Region         string `json:"region,omitempty" yaml:"region,omitempty"`
}
