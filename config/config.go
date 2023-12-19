package config

// TestClientName is string enum for client/user names used in provisioning tests.
type TestClientName string
type PSACT string

const (
	StandardClientName TestClientName = "Standard User"

	RancherPrivileged PSACT = "rancher-privileged"
	RancherRestricted PSACT = "rancher-restricted"
)

var AllRolesNodePool = Nodepool{
	Etcd:         true,
	Controlplane: true,
	Worker:       true,
	Quantity:     1,
}

var EtcdControlPlaneNodePool = Nodepool{
	Etcd:         true,
	Controlplane: true,
	Worker:       false,
	Quantity:     1,
}

var EtcdNodePool = Nodepool{
	Etcd:         true,
	Controlplane: false,
	Worker:       false,
	Quantity:     1,
}

var ControlPlaneNodePool = Nodepool{
	Etcd:         false,
	Controlplane: true,
	Worker:       false,
	Quantity:     1,
}

var WorkerNodePool = Nodepool{
	Etcd:         false,
	Controlplane: false,
	Worker:       true,
	Quantity:     1,
}

var ScaleUpAllRolesNodePool = Nodepool{
	Etcd:         true,
	Controlplane: true,
	Worker:       true,
	Quantity:     4,
}

var ScaleDownAllRolesNodePool = Nodepool{
	Etcd:         true,
	Controlplane: true,
	Worker:       true,
	Quantity:     3,
}

var ScaleUpEtcdControlPlaneNodePool = Nodepool{
	Etcd:         true,
	Controlplane: true,
	Worker:       false,
	Quantity:     3,
}

var ScaleUpEtcdNodePool = Nodepool{
	Etcd:         true,
	Controlplane: false,
	Worker:       false,
	Quantity:     3,
}

var ScaleUpControlPlaneNodePool = Nodepool{
	Etcd:         false,
	Controlplane: true,
	Worker:       false,
	Quantity:     2,
}

var ScaleUpWorkerNodePool = Nodepool{
	Etcd:         false,
	Controlplane: false,
	Worker:       true,
	Quantity:     3,
}

// String stringer for the TestClientName
func (c TestClientName) String() string {
	return string(c)
}

type GoogleAuthEncodedJSON struct {
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url" yaml:"auth_provider_x509_cert_url"`
	AuthURI                 string `json:"auth_uri" yaml:"auth_uri"`
	ClientEmail             string `json:"client_email" yaml:"client_email"`
	ClientID                string `json:"client_id" yaml:"client_id"`
	ClientX509CertURL       string `json:"client_x509_cert_url" yaml:"client_x509_cert_url"`
	PrivateKey              string `json:"private_key" yaml:"private_key"`
	PrivateKeyID            string `json:"private_key_id" yaml:"private_key_id"`
	ProjectID               string `json:"project_id" yaml:"project_id"`
	TokenURI                string `json:"token_uri" yaml:"token_uri"`
	Type                    string `json:"type" yaml:"type"`
}

type Nodepool struct {
	Quantity         int64  `json:"quantity" yaml:"quantity"`
	Etcd             bool   `json:"etcd" yaml:"etcd"`
	Controlplane     bool   `json:"controlplane" yaml:"controlplane"`
	Worker           bool   `json:"worker" yaml:"worker"`
	InstanceType     string `json:"instanceType" yaml:"instanceType"`
	DesiredSize      int64  `json:"desiredSize" yaml:"desiredSize"`
	MaxSize          int64  `json:"maxSize" yaml:"maxSize"`
	MinSize          int64  `json:"minSize" yaml:"minSize"`
	MaxPodsContraint int64  `json:"maxPodsContraint" yaml:"maxPodsContraint"`
}

type TerraformConfig struct {
	Ami                                 string   `json:"ami,omitempty" yaml:"ami,omitempty"`
	AvailabilityZones                   []string `json:"availabilityZones" yaml:"availabilityZones"`
	AWSAccessKey                        string   `json:"awsAccessKey" yaml:"awsAccessKey"`
	AWSInstanceType                     string   `json:"awsInstanceType" yaml:"awsInstanceType"`
	AWSRootSize                         int64    `json:"awsRootSize" yaml:"awsRootSize"`
	AWSSecretKey                        string   `json:"awsSecretKey" yaml:"awsSecretKey"`
	AWSSecurityGroupNames               []string `json:"awsSecurityGroupNames" yaml:"awsSecurityGroupNames"`
	AWSSecurityGroups                   []string `json:"awsSecurityGroups" yaml:"awsSecurityGroups"`
	AWSSubnetID                         string   `json:"awsSubnetID" yaml:"awsSubnetID"`
	AWSSubnets                          []string `json:"awsSubnets" yaml:"awsSubnets"`
	AWSVpcID                            string   `json:"awsVpcID" yaml:"awsVpcID"`
	AWSZoneLetter                       string   `json:"awsZoneLetter" yaml:"awsZoneLetter"`
	AzureClientID                       string   `json:"azureClientID" yaml:"azureClientID"`
	AzureClientSecret                   string   `json:"azureClientSecret" yaml:"azureClientSecret"`
	AzureSubscriptionID                 string   `json:"azureSubscriptionID" yaml:"azureSubscriptionID"`
	CloudCredentialName                 string   `json:"cloudCredentialName" yaml:"cloudCredentialName"`
	DefaultClusterRoleForProjectMembers string   `json:"defaultClusterRoleForProjectMembers" yaml:"defaultClusterRoleForProjectMembers"`
	EnableNetworkPolicy                 bool     `json:"enableNetworkPolicy" yaml:"enableNetworkPolicy"`
	GKENetwork                          string   `json:"gkeNetwork" yaml:"gkeNetwork"`
	GKEProjectID                        string   `json:"gkeProjectID" yaml:"gkeProjectID"`
	GKESubnetwork                       string   `json:"gkeSubnetwork" yaml:"gkeSubnetwork"`
	HostnamePrefix                      string   `json:"hostnamePrefix" yaml:"hostnamePrefix"`
	LinodeImage                         string   `json:"linodeImage" yaml:"linodeImage"`
	LinodeRootPass                      string   `json:"linodeRootPass" yaml:"linodeRootPass"`
	LinodeToken                         string   `json:"linodeToken" yaml:"linodeToken"`
	MachineConfigName                   string   `json:"machineConfigName" yaml:"machineConfigName"`
	Module                              string   `json:"module" yaml:"module"`
	NetworkPlugin                       string   `json:"networkPlugin" yaml:"networkPlugin"`
	NodeTemplateName                    string   `json:"nodeTemplateName" yaml:"nodeTemplateName"`
	OSDiskSizeGB                        int64    `json:"osDiskSizeGB" yaml:"osDiskSizeGB"`
	PrivateAccess                       bool     `json:"privateAccess" yaml:"privateAccess"`
	PublicAccess                        bool     `json:"publicAccess" yaml:"publicAccess"`
	Region                              string   `json:"region" yaml:"region"`
	ResourceGroup                       string   `json:"resourceGroup" yaml:"resourceGroup"`
	ResourceLocation                    string   `json:"resourceLocation" yaml:"resourceLocation"`
	VMSize                              string   `json:"vmSize" yaml:"vmSize"`
}

type TerratestConfig struct {
	KubernetesVersion         string     `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	NodeCount                 int64      `json:"nodeCount" yaml:"nodeCount"`
	Nodepools                 []Nodepool `json:"nodepools" yaml:"nodepools"`
	ScaledDownNodeCount       int64      `json:"scaledDownNodeCount" yaml:"scaledDownNodeCount"`
	ScaledDownNodepools       []Nodepool `json:"scaledDownNodepools" yaml:"scaledDownNodepools"`
	ScaledUpNodeCount         int64      `json:"scaledUpNodeCount" yaml:"scaledUpNodeCount"`
	ScaledUpNodepools         []Nodepool `json:"scaledUpNodepools" yaml:"scaledUpNodepools"`
	UpgradedKubernetesVersion string     `json:"upgradedKubernetesVersion" yaml:"upgradedKubernetesVersion"`
	PSACT                     string     `json:"psact" yaml:"psact"`
}
