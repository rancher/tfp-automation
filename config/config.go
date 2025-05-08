package config

import (
	"errors"
	"fmt"
	"path"
	"runtime"

	"github.com/imdario/mergo"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	shepherdConfig "github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/config/operations"
	"github.com/rancher/tfp-automation/config/authproviders"
	aws "github.com/rancher/tfp-automation/config/nodeproviders/aws"
	azure "github.com/rancher/tfp-automation/config/nodeproviders/azure"
	google "github.com/rancher/tfp-automation/config/nodeproviders/google"
	harvester "github.com/rancher/tfp-automation/config/nodeproviders/harvester"
	linode "github.com/rancher/tfp-automation/config/nodeproviders/linode"
	vsphere "github.com/rancher/tfp-automation/config/nodeproviders/vsphere"
	"github.com/rancher/tfp-automation/defaults/configs"
)

type TestClientName string
type Role string
type PSACT string

const (
	TerraformConfigurationFileKey = "terraform"
	TerratestConfigurationFileKey = "terratest"

	AdminClientName    TestClientName = "Admin User"
	StandardClientName TestClientName = "Standard User"

	ClusterOwner Role = "cluster-owner"
	ProjectOwner Role = "project-owner"

	RancherPrivileged PSACT = "rancher-privileged"
	RancherRestricted PSACT = "rancher-restricted"

	defaultFilename      = "defaults.yaml"
	provisioningFilename = "provisioning.yaml"
)

var EtcdNodePool = Nodepool{
	Etcd:         true,
	Controlplane: false,
	Worker:       false,
	Quantity:     3,
}

var ControlPlaneNodePool = Nodepool{
	Etcd:         false,
	Controlplane: true,
	Worker:       false,
	Quantity:     2,
}

var WorkerNodePool = Nodepool{
	Etcd:         false,
	Controlplane: false,
	Worker:       true,
	Quantity:     3,
}

var AllRolesNodePool = Nodepool{
	Etcd:         true,
	Controlplane: true,
	Worker:       true,
	Quantity:     1,
}

var ScaleUpEtcdNodePool = Nodepool{
	Etcd:         true,
	Controlplane: false,
	Worker:       false,
	Quantity:     5,
}

var ScaleUpControlPlaneNodePool = Nodepool{
	Etcd:         false,
	Controlplane: true,
	Worker:       false,
	Quantity:     3,
}

var ScaleUpWorkerNodePool = Nodepool{
	Etcd:         false,
	Controlplane: false,
	Worker:       true,
	Quantity:     5,
}

// String stringer for the TestClientName
func (c TestClientName) String() string {
	return string(c)
}

type Nodepool struct {
	Quantity         int64  `json:"quantity,omitempty" yaml:"quantity,omitempty"`
	Etcd             bool   `json:"etcd,omitempty" yaml:"etcd,omitempty"`
	Controlplane     bool   `json:"controlplane,omitempty" yaml:"controlplane,omitempty"`
	Worker           bool   `json:"worker,omitempty" yaml:"worker,omitempty"`
	InstanceType     string `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`
	DesiredSize      int64  `json:"desiredSize,omitempty" yaml:"desiredSize,omitempty"`
	MaxSize          int64  `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	MinSize          int64  `json:"minSize,omitempty" yaml:"minSize,omitempty"`
	MaxPodsContraint int64  `json:"maxPodsContraint,omitempty" yaml:"maxPodsContraint,omitempty"`
}

type Proxy struct {
	ProxyBastion string `json:"proxyBastion,omitempty" yaml:"proxyBastion,omitempty"`
}

type PrivateRegistries struct {
	AuthConfigSecretName   string `json:"authConfigSecretName,omitempty" yaml:"authConfigSecretName,omitempty"`
	CABundle               string `json:"caBundle,omitempty" yaml:"caBundle,omitempty"`
	EngineInsecureRegistry string `json:"engineInsecureRegistry,omitempty" yaml:"engineInsecureRegistry,omitempty"`
	Insecure               bool   `json:"insecure,omitempty" yaml:"insecure,omitempty"`
	Password               string `json:"password,omitempty" yaml:"password,omitempty"`
	SystemDefaultRegistry  string `json:"systemDefaultRegistry,omitempty" yaml:"systemDefaultRegistry,omitempty"`
	TLSSecretName          string `json:"tlsSecretName,omitempty" yaml:"tlsSecretName,omitempty"`
	URL                    string `json:"url,omitempty" yaml:"url,omitempty"`
	Username               string `json:"username,omitempty" yaml:"username,omitempty"`
}

type Standalone struct {
	AirgapInternalFQDN             string `json:"airgapInternalFQDN,omitempty" yaml:"airgapInternalFQDN,omitempty"`
	BootstrapPassword              string `json:"bootstrapPassword,omitempty" yaml:"bootstrapPassword,omitempty"`
	CertManagerVersion             string `json:"certManagerVersion,omitempty" yaml:"certManagerVersion,omitempty"`
	K3SVersion                     string `json:"k3sVersion,omitempty" yaml:"k3sVersion,omitempty"`
	RancherAgentImage              string `json:"rancherAgentImage,omitempty" yaml:"rancherAgentImage,omitempty"`
	RancherChartRepository         string `json:"rancherChartRepository,omitempty" yaml:"rancherChartRepository,omitempty"`
	RancherHostname                string `json:"rancherHostname,omitempty" yaml:"rancherHostname,omitempty"`
	RancherImage                   string `json:"rancherImage,omitempty" yaml:"rancherImage,omitempty"`
	RancherTagVersion              string `json:"rancherTagVersion,omitempty" yaml:"rancherTagVersion,omitempty"`
	Repo                           string `json:"repo,omitempty" yaml:"repo,omitempty"`
	OSUser                         string `json:"osUser,omitempty" yaml:"osUser,omitempty"`
	OSGroup                        string `json:"osGroup,omitempty" yaml:"osGroup,omitempty"`
	RKE2Version                    string `json:"rke2Version,omitempty" yaml:"rke2Version,omitempty"`
	UpgradeAirgapRancher           bool   `json:"upgradeAirgapRancher,omitempty" yaml:"upgradeAirgapRancher,omitempty"`
	UpgradeProxyRancher            bool   `json:"upgradeProxyRancher,omitempty" yaml:"upgradeProxyRancher,omitempty"`
	UpgradeRancher                 bool   `json:"upgradeRancher,omitempty" yaml:"upgradeRancher,omitempty"`
	UpgradedRancherChartRepository string `json:"upgradedRancherChartRepository,omitempty" yaml:"upgradedRancherChartRepository,omitempty"`
	UpgradedRancherImage           string `json:"upgradedRancherImage,omitempty" yaml:"upgradedRancherImage,omitempty"`
	UpgradedRancherAgentImage      string `json:"upgradedRancherAgentImage,omitempty" yaml:"upgradedRancherAgentImage,omitempty"`
	UpgradedRancherRepo            string `json:"upgradedRancherRepo,omitempty" yaml:"upgradedRancherRepo,omitempty"`
	UpgradedRancherTagVersion      string `json:"upgradedRancherTagVersion,omitempty" yaml:"upgradedRancherTagVersion,omitempty"`
}

type StandaloneRegistry struct {
	AssetsPath         string `json:"assetsPath,omitempty" yaml:"assetsPath,omitempty"`
	Authenticated      bool   `json:"authenticated,omitempty" yaml:"authenticated,omitempty"`
	ECRAMI             string `json:"ecrAMI,omitempty" yaml:"ecrAMI,omitempty"`
	ECRURI             string `json:"ecrURI,omitempty" yaml:"ecrURI,omitempty"`
	ECRUsername        string `json:"ecrUsername,omitempty" yaml:"ecrUsername,omitempty"`
	ECRPassword        string `json:"ecrPassword,omitempty" yaml:"ecrPassword,omitempty"`
	RegistryName       string `json:"registryName,omitempty" yaml:"registryName,omitempty"`
	RegistryPassword   string `json:"registryPassword,omitempty" yaml:"registryPassword,omitempty"`
	RegistryUsername   string `json:"registryUsername,omitempty" yaml:"registryUsername,omitempty"`
	UpgradedAssetsPath string `json:"upgradedAssetsPath,omitempty" yaml:"upgradedAssetsPath,omitempty"`
}

type TerraformConfig struct {
	AWSConfig                           aws.Config                   `json:"awsConfig,omitempty" yaml:"awsConfig,omitempty"`
	AWSCredentials                      aws.Credentials              `json:"awsCredentials,omitempty" yaml:"awsCredentials,omitempty"`
	AzureConfig                         azure.Config                 `json:"azureConfig,omitempty" yaml:"azureConfig,omitempty"`
	AzureCredentials                    azure.Credentials            `json:"azureCredentials,omitempty" yaml:"azureCredentials,omitempty"`
	GoogleConfig                        google.Config                `json:"googleConfig,omitempty" yaml:"googleConfig,omitempty"`
	GoogleCredentials                   google.Credentials           `json:"googleCredentials,omitempty" yaml:"googleCredentials,omitempty"`
	HarvesterConfig                     harvester.Config             `json:"harvesterConfig,omitempty" yaml:"harvesterConfig,omitempty"`
	HarvesterCredentials                harvester.Credentials        `json:"harvesterCredentials,omitempty" yaml:"harvesterCredentials,omitempty"`
	LinodeConfig                        linode.Config                `json:"linodeConfig,omitempty" yaml:"linodeConfig,omitempty"`
	LinodeCredentials                   linode.Credentials           `json:"linodeCredentials,omitempty" yaml:"linodeCredentials,omitempty"`
	VsphereConfig                       vsphere.Config               `json:"vsphereConfig,omitempty" yaml:"vsphereConfig,omitempty"`
	VsphereCredentials                  vsphere.Credentials          `json:"vsphereCredentials,omitempty" yaml:"vsphereCredentials,omitempty"`
	ADConfig                            authproviders.ADConfig       `json:"adConfig,omitempty" yaml:"adConfig,omitempty"`
	AzureADConfig                       authproviders.AzureADConfig  `json:"azureADConfig,omitempty" yaml:"azureADConfig,omitempty"`
	GithubConfig                        authproviders.GithubConfig   `json:"githubConfig,omitempty" yaml:"githubConfig,omitempty"`
	OktaConfig                          authproviders.OktaConfig     `json:"oktaConfig,omitempty" yaml:"oktaConfig,omitempty"`
	OpenLDAPConfig                      authproviders.OpenLDAPConfig `json:"openLDAPConfig,omitempty" yaml:"openLDAPConfig,omitempty"`
	AuthProvider                        string                       `json:"authProvider,omitempty" yaml:"authProvider,omitempty"`
	ResourcePrefix                      string                       `json:"resourcePrefix,omitempty" yaml:"resourcePrefix,omitempty"`
	CNI                                 string                       `json:"cni,omitempty" yaml:"cni,omitempty"`
	ChartValues                         string                       `json:"chartValues,omitempty" yaml:"chartValues,omitempty"`
	DisableKubeProxy                    string                       `json:"disable-kube-proxy,omitempty" yaml:"disable-kube-proxy,omitempty"`
	DefaultClusterRoleForProjectMembers string                       `json:"defaultClusterRoleForProjectMembers,omitempty" yaml:"defaultClusterRoleForProjectMembers,omitempty"`
	EnableNetworkPolicy                 bool                         `json:"enableNetworkPolicy,omitempty" yaml:"enableNetworkPolicy,omitempty"`
	ETCD                                *rkev1.ETCD                  `json:"etcd,omitempty" yaml:"etcd,omitempty"`
	ETCDRKE1                            *management.ETCDService      `json:"etcdRKE1,omitempty" yaml:"etcdRKE1,omitempty"`
	Module                              string                       `json:"module,omitempty" yaml:"module,omitempty"`
	NetworkPlugin                       string                       `json:"networkPlugin,omitempty" yaml:"networkPlugin,omitempty"`
	PrivateKeyPath                      string                       `json:"privateKeyPath,omitempty" yaml:"privateKeyPath,omitempty"`
	PrivateRegistries                   *PrivateRegistries           `json:"privateRegistries,omitempty" yaml:"privateRegistries,omitempty"`
	Proxy                               *Proxy                       `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	Provider                            string                       `json:"provider,omitempty" yaml:"provider,omitempty"`
	Standalone                          *Standalone                  `json:"standalone,omitempty" yaml:"standalone,omitempty"`
	StandaloneRegistry                  *StandaloneRegistry          `json:"standaloneRegistry,omitempty" yaml:"standaloneRegistry,omitempty"`
	TimeSleep                           string                       `json:"timeSleep,omitempty" yaml:"timeSleep,omitempty"`
	WindowsPrivateKeyPath               string                       `json:"windowsPrivateKeyPath,omitempty" yaml:"windowsPrivateKeyPath,omitempty"`
}

type Scaling struct {
	ScaledDownNodeCount int64      `json:"scaledDownNodeCount,omitempty" yaml:"scaledDownNodeCount,omitempty"`
	ScaledDownNodepools []Nodepool `json:"scaledDownNodepools,omitempty" yaml:"scaledDownNodepools,omitempty"`
	ScaledUpNodeCount   int64      `json:"scaledUpNodeCount,omitempty" yaml:"scaledUpNodeCount,omitempty"`
	ScaledUpNodepools   []Nodepool `json:"scaledUpNodepools,omitempty" yaml:"scaledUpNodepools,omitempty"`
}

type Snapshots struct {
	CreateSnapshot  bool   `json:"createSnapshot,omitempty" yaml:"createSnapshot,omitempty"`
	RestoreSnapshot bool   `json:"restoreSnapshot,omitempty" yaml:"restoreSnapshot,omitempty"`
	SnapshotName    string `json:"snapshotName,omitempty" yaml:"snapshotName,omitempty"`
	SnapshotRestore string `json:"snapshotRestore,omitempty" yaml:"snapshotRestore,omitempty"`
}

type TerratestConfig struct {
	KubernetesVersion         string     `json:"kubernetesVersion,omitempty" yaml:"kubernetesVersion,omitempty"`
	LocalQaseReporting        bool       `json:"localQaseReporting,omitempty" yaml:"localQaseReporting,omitempty" default:"false"`
	NodeCount                 int64      `json:"nodeCount,omitempty" yaml:"nodeCount,omitempty"`
	Nodepools                 []Nodepool `json:"nodepools,omitempty" yaml:"nodepools,omitempty"`
	PSACT                     string     `json:"psact,omitempty" yaml:"psact,omitempty"`
	ScalingInput              Scaling    `json:"scalingInput,omitempty" yaml:"scalingInput,omitempty"`
	SnapshotInput             Snapshots  `json:"snapshotInput,omitempty" yaml:"snapshotInput,omitempty"`
	StandaloneLogging         bool       `json:"standaloneLogging,omitempty" yaml:"standaloneLogging,omitempty"`
	TFLogging                 bool       `json:"tfLogging,omitempty" yaml:"tfLogging,omitempty"`
	UpgradedKubernetesVersion string     `json:"upgradedKubernetesVersion,omitempty" yaml:"upgradedKubernetesVersion,omitempty"`
	WindowsNodeCount          int64      `json:"windowsNodeCount,omitempty" yaml:"windowsNodeCount,omitempty"`
}

// LoadTFPConfigs loads the TFP configurations from the provided map
func LoadTFPConfigs(cattleConfig map[string]any) (*rancher.Config, *TerraformConfig, *TerratestConfig) {
	rancherConfig := new(rancher.Config)
	operations.LoadObjectFromMap(configs.Rancher, cattleConfig, rancherConfig)

	terraformConfig := new(TerraformConfig)
	operations.LoadObjectFromMap(TerraformConfigurationFileKey, cattleConfig, terraformConfig)

	terratestConfig := new(TerratestConfig)
	operations.LoadObjectFromMap(TerratestConfigurationFileKey, cattleConfig, terratestConfig)

	return rancherConfig, terraformConfig, terratestConfig
}

// LoadPackageDefaults loads the specified filename in the same package as the test
func LoadPackageDefaults(cattleConfig map[string]any, filePath string) (map[string]any, error) {
	if filePath == "" {
		filePath = defaultFilename
	}

	defaultsConfig := shepherdConfig.LoadConfigFromFile(filePath)
	err := mergo.Merge(&defaultsConfig, cattleConfig, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	return defaultsConfig, nil
}

// LoadProvisioningDefaults loads the provisioning.yaml file if it exists and merges values provided in the cattleConfig
func LoadProvisioningDefaults(cattleConfig map[string]any, filename string) (map[string]any, error) {
	if filename == "" {
		filename = provisioningFilename
	}

	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		err := errors.New(fmt.Sprintf("Unable to locate directory of %s", filename))
		return nil, err
	}

	configPath := path.Dir(filePath)

	defaultsConfig := shepherdConfig.LoadConfigFromFile(configPath + "/" + filename)
	err := mergo.Merge(&defaultsConfig, cattleConfig, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	return defaultsConfig, nil
}
