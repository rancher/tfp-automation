package google

type Config struct {
	Image       string `json:"image,omitempty" yaml:"image,omitempty"`
	KeyPath     string `json:"keyPath,omitempty" yaml:"keyPath,omitempty"`
	MachineType string `json:"machineType,omitempty" yaml:"machineType,omitempty"`
	Network     string `json:"network,omitempty" yaml:"network,omitempty"`
	ProjectID   string `json:"projectID,omitempty" yaml:"projectID,omitempty"`
	Size        int64  `json:"size,omitempty" yaml:"size,omitempty"`
	SSHUser     string `json:"sshUser,omitempty" yaml:"sshUser,omitempty"`
	Subnetwork  string `json:"subnetwork,omitempty" yaml:"subnetwork,omitempty"`
	Region      string `json:"region,omitempty" yaml:"region,omitempty"`
	Zone        string `json:"zone,omitempty" yaml:"zone,omitempty"`
}
