package linode

type Config struct {
	ClientConnThrottle int64    `json:"clientConnThrottle,omitempty" yaml:"clientConnThrottle,omitempty"`
	Domain             string   `json:"domain,omitempty" yaml:"domain,omitempty"`
	LinodeImage        string   `json:"linodeImage,omitempty" yaml:"linodeImage,omitempty"`
	LinodeRootPass     string   `json:"linodeRootPass,omitempty" yaml:"linodeRootPass,omitempty"`
	PrivateIP          bool     `json:"privateIP,omitempty" yaml:"privateIP,omitempty"`
	Region             string   `json:"region,omitempty" yaml:"region,omitempty"`
	SOAEmail           string   `json:"soaEmail,omitempty" yaml:"soaEmail,omitempty"`
	SwapSize           int64    `json:"swapSize,omitempty" yaml:"swapSize,omitempty"`
	Tags               []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Timeout            string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Type               string   `json:"type,omitempty" yaml:"type,omitempty"`
}
