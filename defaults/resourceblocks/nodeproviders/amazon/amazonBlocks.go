package amazon

const (
	EC2CredentialConfig = "amazonec2_credential_config"
	DefaultRegion       = "default_region"

	EC2Config = "amazonec2_config"
	EKSConfig = "eks_config_v2"

	Subnets        = "subnets"
	SecurityGroups = "security_groups"
	PrivateAccess  = "private_access"
	PublicAccess   = "public_access"

	AMI           = "ami"
	SecurityGroup = "security_group"
	SubnetID      = "subnet_id"
	VPCID         = "vpc_id"
	Zone          = "zone"
	RootSize      = "root_size"

	NodeGroups   = "node_groups"
	DiskSize     = "disk_size"
	InstanceType = "instance_type"
	VolumeType   = "volume_type"
	SSHUser      = "ssh_user"
	DesiredSize  = "desired_size"
	MaxSize      = "max_size"
	MinSize      = "min_size"

	EnablePrimaryIPv6 = "enable_primary_ipv6"
	HTTPProtocolIPv6  = "http_protocol_ipv6"
	IPv6AddressCount  = "ipv6_address_count"
	IPv6AddressOnly   = "ipv6_address_only"
)
