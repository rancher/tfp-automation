package provisioning

const (
	resource     = "resource"
	resourceKind = "kind"
	resourceName = "name"

	dependsOn    = "depends_on"
	generateName = "generate_name"

	defaultPodSecurityAdmission = "default_pod_security_admission_configuration_template_name"
	podSecurityAdmission        = "rancher2_pod_security_admission_configuration_template"
	cloudCredential             = "rancher2_cloud_credential"
	cluster                     = "rancher2_cluster"

	rkeConfig                           = "rke_config"
	kubernetesVersion                   = "kubernetes_version"
	network                             = "network"
	plugin                              = "plugin"
	services                            = "services"
	enableNetworkPolicy                 = "enable_network_policy"
	defaultClusterRoleForProjectMembers = "default_cluster_role_for_project_members"

	accessKey         = "access_key"
	secretKey         = "secret_key"
	cloudCredentialID = "cloud_credential_id"

	machineConfig    = "machine_config"
	machinePools     = "machine_pools"
	nodePool         = "pool"
	etcd             = "etcd"
	rancherClusterID = "cluster_id"
	quantity         = "quantity"

	endpoint = "endpoint"
	folder   = "folder"
	region   = "region"
)
