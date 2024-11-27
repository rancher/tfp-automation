package defaults

const (
	Data         = "data"
	Resource     = "resource"
	ResourceKind = "kind"
	ResourceName = "name"

	DependsOn    = "depends_on"
	GenerateName = "generate_name"
	Defaults     = "defaults"
	Index        = "index"
	File         = "file"
	Custom       = "custom"

	Rancher2Source      = "rancher/rancher2"
	Rancher2LocalSource = "terraform.local/local/rancher2"

	DefaultPodSecurityAdmission = "default_pod_security_admission_configuration_template_name"
	PodSecurityAdmission        = "rancher2_pod_security_admission_configuration_template"
	CloudCredential             = "rancher2_cloud_credential"
	Cluster                     = "rancher2_cluster"
	ClusterV2                   = "rancher2_cluster_v2"

	RkeConfig                           = "rke_config"
	KubernetesVersion                   = "kubernetes_version"
	Network                             = "network"
	Plugin                              = "plugin"
	RKE1PrivateRegistries               = "private_registries"
	PrivateRegistries                   = "registries"
	EngineInsecureRegistry              = "engine_insecure_registry"
	Config                              = "config"
	Configs                             = "configs"
	Mirrors                             = "mirrors"
	Services                            = "services"
	EnableNetworkPolicy                 = "enable_network_policy"
	DefaultClusterRoleForProjectMembers = "default_cluster_role_for_project_members"
	RancherBaseline                     = "rancher-baseline"

	AccessKey         = "access_key"
	SecretKey         = "secret_key"
	CloudCredentialID = "cloud_credential_id"

	MachineConfig         = "machine_config"
	MachineGlobalConfig   = "machine_global_config"
	MachineSelectorConfig = "machine_selector_config"
	MachinePools          = "machine_pools"
	NodePool              = "auto-tfp-pool"
	Etcd                  = "etcd"
	Enabled               = "enabled"
	RancherClusterID      = "cluster_id"
	Quantity              = "quantity"

	Endpoint = "endpoint"
	Folder   = "folder"
	Region   = "region"

	Ami                 = "ami"
	Count               = "count"
	InstanceType        = "instance_type"
	SubnetId            = "subnet_id"
	VpcId               = "vpc_id"
	VpcSecurityGroupIds = "vpc_security_group_ids"
	KeyName             = "key_name"
	AwsInstance         = "aws_instance"
	Name                = "Name"
	RootBlockDevice     = "root_block_device"
	Tags                = "tags"
	VolumeSize          = "volume_size"
	Timeout             = "timeout"

	Locals                   = "locals"
	RoleFlags                = "role_flags"
	EtcdRoleFlag             = "--etcd"
	ControlPlaneRoleFlag     = "--controlplane"
	WorkerRoleFlag           = "--worker"
	OriginalNodeCommand      = "original_node_command"
	InsecureNodeCommand      = "insecure_node_command"
	ClusterRegistrationToken = "cluster_registration_token"
	NodeCommand              = "node_command"

	ConfigPath    = "config_path"
	Connection    = "connection"
	Host          = "host"
	Inline        = "inline"
	NullResource  = "null_resource"
	RegisterNodes = "register_nodes"
	PrivateKey    = "private_key"
	Provisioner   = "provisioner"
	RemoteExec    = "remote-exec"
	Ssh           = "ssh"
	Type          = "type"
	User          = "user"
	Self          = "self"
	PublicIp      = "public_ip"
	Length        = "length"

	ApiUrl            = "api_url"
	Aws               = "aws"
	Helm              = "helm"
	Local             = "local"
	Kubernetes        = "kubernetes"
	AwsSource         = "hashicorp/aws"
	HelmSource        = "hashicorp/helm"
	Insecure          = "insecure"
	LocalSource       = "hashicorp/local"
	Provider          = "provider"
	Rancher2          = "rancher2"
	Rc                = "-rc"
	RequiredProviders = "required_providers"
	Source            = "source"
	Terraform         = "terraform"
	TokenKey          = "token_key"
	Version           = "version"

	DefaultAction                     = "default_action"
	HealthCheck                       = "health_check"
	LoadBalancer                      = "aws_lb"
	LoadBalancerARN                   = "load_balancer_arn"
	LoadBalancerListener              = "aws_lb_listener"
	LoadBalancerType                  = "load_balancer_type"
	LoadBalancerTargetGroup           = "aws_lb_target_group"
	LoadBalancerTargetGroupAttachment = "aws_lb_target_group_attachment"
	Port                              = "port"
	Route53Record                     = "aws_route53_record"
	Route53Zone                       = "aws_route53_zone"
	Subnets                           = "subnets"
	TargetGroup80Attachment           = "aws_tg_attachment_80_server"
	TargetGroup443Attachment          = "aws_tg_attachment_443_server"
	TargetGroup6443Attachment         = "aws_tg_attachment_6443_server"
	TargetGroup9345Attachment         = "aws_tg_attachment_9345_server"
	TargetGroupARN                    = "target_group_arn"
	TargetGroupPrefix                 = "aws_tg_"
	TargetID                          = "target_id"
)
