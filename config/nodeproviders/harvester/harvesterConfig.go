package harvester

type Config struct {
	DiskSize     string   `json:"diskSize,omitempty" yaml:"diskSize,omitempty"`
	CPUCount     string   `json:"cpuCount,omitempty" yaml:"cpuCount,omitempty"`
	MemorySize   string   `json:"memorySize,omitempty" yaml:"memorySize,omitempty"`
	NetworkNames []string `json:"networkNames,omitempty" yaml:"networkNames,omitempty"`
	ImageName    string   `json:"imageName,omitempty" yaml:"imageName,omitempty"`
	SSHUser      string   `json:"sshUser,omitempty" yaml:"sshUser,omitempty"`
	VMNamespace  string   `json:"vmNamespace,omitempty" yaml:"vmNamespace,omitempty"`
	UserData     string   `json:"userData,omitempty" yaml:"userData,omitempty"`
}
