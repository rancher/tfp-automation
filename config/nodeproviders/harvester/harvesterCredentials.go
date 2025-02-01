package harvester

type Credentials struct {
	ClusterID         string `json:"clusterID,omitempty" yaml:"clusterID,omitempty"`
	ClusterType       string `json:"clusterType,omitempty" yaml:"clusterType,omitempty"`
	KubeconfigContent string `json:"kubeconfigContent,omitempty" yaml:"kubeconfigContent,omitempty"`
}
