package vsphere

type Credentials struct {
	Password    string `json:"password,omitempty" yaml:"password,omitempty"`
	Username    string `json:"username,omitempty" yaml:"username,omitempty"`
	Vcenter     string `json:"vcenter,omitempty" yaml:"vcenter,omitempty"`
	VcenterPort string `json:"vcenterPort,omitempty" yaml:"vcenterPort,omitempty"`
}
