package resourceblocks

const (
	Resource     = "resource"
	ResourceKind = "kind"
	ResourceName = "name"

	DependsOn    = "depends_on"
	GenerateName = "generate_name"

	DefaultPodSecurityAdmission = "default_pod_security_admission_configuration_template_name"
	PodSecurityAdmission        = "rancher2_pod_security_admission_configuration_template"
	CloudCredential             = "rancher2_cloud_credential"
	Cluster                     = "rancher2_cluster"

	RKEConfig                           = "rke_config"
	KubernetesVersion                   = "kubernetes_version"
	Network                             = "network"
	Plugin                              = "plugin"
	Services                            = "services"
	EnableNetworkPolicy                 = "enable_network_policy"
	DefaultClusterRoleForProjectMembers = "default_cluster_role_for_project_members"

	AccessKey         = "access_key"
	SecretKey         = "secret_key"
	CloudCredentialID = "cloud_credential_id"

	MachineConfig = "machine_config"
	MachinePools  = "machine_pools"
	Pool          = "pool"
	Etcd          = "etcd"
	ClusterID     = "cluster_id"
	Quantity      = "quantity"

	Endpoint = "endpoint"
	Folder   = "folder"
	Region   = "region"
)
