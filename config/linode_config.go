package config

type LinodeConfig struct {
	LinodeImage    string `json:"linodeImage,omitempty" yaml:"linodeImage,omitempty"`
	LinodeRootPass string `json:"linodeRootPass,omitempty" yaml:"linodeRootPass,omitempty"`
	LinodeToken    string `json:"linodeToken,omitempty" yaml:"linodeToken,omitempty"`
	Region         string `json:"region,omitempty" yaml:"region,omitempty"`
}
