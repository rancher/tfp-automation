package config

type AzureConfig struct {
	AvailabilityZones   []string `json:"availabilityZones,omitempty" yaml:"availabilityZones,omitempty"`
	AzureClientID       string   `json:"azureClientID,omitempty" yaml:"azureClientID,omitempty"`
	AzureClientSecret   string   `json:"azureClientSecret,omitempty" yaml:"azureClientSecret,omitempty"`
	AzureSubscriptionID string   `json:"azureSubscriptionID,omitempty" yaml:"azureSubscriptionID,omitempty"`
	OSDiskSizeGB        int64    `json:"osDiskSizeGB,omitempty" yaml:"osDiskSizeGB,omitempty"`
	ResourceGroup       string   `json:"resourceGroup,omitempty" yaml:"resourceGroup,omitempty"`
	ResourceLocation    string   `json:"resourceLocation,omitempty" yaml:"resourceLocation,omitempty"`
	VMSize              string   `json:"vmSize,omitempty" yaml:"vmSize,omitempty"`
}
