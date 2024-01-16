package nodeproviders

type AzureConfig struct {
	AvailabilitySet   string   `json:"availabilitySet,omitempty" yaml:"availabilitySet,omitempty"`
	AvailabilityZones []string `json:"availabilityZones,omitempty" yaml:"availabilityZones,omitempty"`
	ClientID          string   `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	ClientSecret      string   `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	CustomData        string   `json:"customData,omitempty" yaml:"customData,omitempty"`
	DiskSize          string   `json:"diskSize,omitempty" yaml:"diskSize,omitempty"`
	DNS               string   `json:"dns,omitempty" yaml:"dns,omitempty"`
	Environment       string   `json:"environment,omitempty" yaml:"environment,omitempty"`
	FaultDomainCount  string   `json:"faultDomainCount,omitempty" yaml:"faultDomainCount,omitempty"`
	Image             string   `json:"image,omitempty" yaml:"image,omitempty"`
	Location          string   `json:"location,omitempty" yaml:"location,omitempty"`
	ManagedDisks      bool     `json:"managedDisks,omitempty" yaml:"managedDisks,omitempty"`
	NoPublicIP        bool     `json:"noPublicIp,omitempty" yaml:"noPublicIp,omitempty"`
	NSG               string   `json:"nsg,omitempty" yaml:"nsg,omitempty"`
	OpenPort          []string `json:"openPort,omitempty" yaml:"openPort,omitempty"`
	OSDiskSizeGB      int64    `json:"osDiskSizeGB,omitempty" yaml:"osDiskSizeGB,omitempty"`
	PrivateIPAddress  string   `json:"privateIpAddress,omitempty" yaml:"privateIpAddress,omitempty"`
	ResourceGroup     string   `json:"resourceGroup,omitempty" yaml:"resourceGroup,omitempty"`
	ResourceLocation  string   `json:"resourceLocation,omitempty" yaml:"resourceLocation,omitempty"`
	Size              string   `json:"size,omitempty" yaml:"size,omitempty"`
	SSHUser           string   `json:"sshUser,omitempty" yaml:"sshUser,omitempty"`
	StaticPublicIP    bool     `json:"staticPublicIp,omitempty" yaml:"staticPublicIp,omitempty"`
	StorageType       string   `json:"storageType,omitempty" yaml:"storageType,omitempty"`
	Subnet            string   `json:"subnet,omitempty" yaml:"subnet,omitempty"`
	SubnetPrefix      string   `json:"subnetPrefix,omitempty" yaml:"subnetPrefix,omitempty"`
	SubscriptionID    string   `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	TenantID          string   `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	UpdateDomainCount string   `json:"updateDomainCount,omitempty" yaml:"updateDomainCount,omitempty"`
	UsePrivateIP      bool     `json:"usePrivateIp,omitempty" yaml:"usePrivateIp,omitempty"`
	VMSize            string   `json:"vmSize,omitempty" yaml:"vmSize,omitempty"`
	Vnet              string   `json:"vnet,omitempty" yaml:"vnet,omitempty"`
}
