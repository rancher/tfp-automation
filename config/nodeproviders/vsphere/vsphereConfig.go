package vsphere

type Config struct {
	Boot2dockerURL         string   `json:"boot2dockerURL,omitempty" yaml:"boot2dockerURL,omitempty"`
	Cfgparam               []string `json:"cfgparam,omitempty" yaml:"cfgparam,omitempty"`
	CloneFrom              string   `json:"cloneFrom,omitempty" yaml:"cloneFrom,omitempty"`
	CloudConfig            string   `json:"cloudConfigv" yaml:"cloudConfig,omitempty"`
	Cloudinit              string   `json:"cloudinit,omitempty" yaml:"cloudinit,omitempty"`
	ContentLibrary         string   `json:"contentLibrary,omitempty" yaml:"contentLibrary,omitempty"`
	CPUCount               string   `json:"cpuCount,omitempty" yaml:"cpuCount,omitempty"`
	CreationType           string   `json:"creationType,omitempty" yaml:"creationType,omitempty"`
	CustomAttribute        []string `json:"customAttribute,omitempty" yaml:"customAttribute,omitempty"`
	DataCenter             string   `json:"dataCenter,omitempty" yaml:"dataCenter,omitempty"`
	DataStore              string   `json:"dataStore,omitempty" yaml:"dataStore,omitempty"`
	DatastoreCluster       string   `json:"datastoreCluster,omitempty" yaml:"datastoreCluster,omitempty"`
	DiskSize               string   `json:"diskSize,omitempty" yaml:"diskSize,omitempty"`
	Folder                 string   `json:"folder,omitempty" yaml:"folder,omitempty"`
	HostSystem             string   `json:"hostSystem,omitempty" yaml:"hostSystem,omitempty"`
	MemorySize             string   `json:"memorySize,omitempty" yaml:"memorySize,omitempty"`
	Network                []string `json:"network,omitempty" yaml:"network,omitempty"`
	OS                     string   `json:"os,omitempty" yaml:"os,omitempty"`
	Pool                   string   `json:"pool,omitempty" yaml:"pool,omitempty"`
	SSHPassword            string   `json:"sshPassword,omitempty" yaml:"sshPassword,omitempty"`
	SSHPort                string   `json:"sshPort,omitempty" yaml:"sshPort,omitempty"`
	SSHUser                string   `json:"sshUser,omitempty" yaml:"sshUser,omitempty"`
	SSHUserGroup           string   `json:"sshUserGroup,omitempty" yaml:"sshUserGroup,omitempty"`
	Tag                    []string `json:"tag,omitempty" yaml:"tag,omitempty"`
	VappIpallocationpolicy string   `json:"vappIpallocationpolicy,omitempty" yaml:"vappIpallocationpolicy,omitempty"`
	VappIpprotocol         string   `json:"vappIpprotocol,omitempty" yaml:"vappIpprotocol,omitempty"`
	VappProperty           []string `json:"vappProperty,omitempty" yaml:"vappProperty,omitempty"`
	VappTransport          string   `json:"vappTransport,omitempty" yaml:"vappTransport,omitempty"`
}
