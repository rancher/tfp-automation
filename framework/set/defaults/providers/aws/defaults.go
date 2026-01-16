package aws

const (
	AwsSource = "hashicorp/aws"

	AccessKey = "access_key"
	SecretKey = "secret_key"
	Endpoint  = "endpoint"
	Folder    = "folder"
	Region    = "region"

	Ami                      = "ami"
	AssociatePublicIPAddress = "associate_public_ip_address"
	EnablePrimaryIPv6        = "enable_primary_ipv6"
	InstanceType             = "instance_type"
	IPAddressType            = "ip_address_type"
	IPv6                     = "ipv6"
	IPV6AddressCount         = "ipv6_address_count"
	SecurityGroups           = "security_groups"
	SubnetId                 = "subnet_id"
	TargetType               = "target_type"
	VpcId                    = "vpc_id"
	VpcSecurityGroupIds      = "vpc_security_group_ids"
	KeyName                  = "key_name"
	Aws                      = "aws"
	AwsInstance              = "aws_instance"
	Name                     = "Name"
	Nodes                    = "nodes"
	RootBlockDevice          = "root_block_device"
	VolumeSize               = "volume_size"
	Timeout                  = "timeout"

	InternalLoadBalancer = "aws_internal_lb"
	LoadBalancer         = "aws_lb"

	LoadBalancerARN               = "load_balancer_arn"
	LoadBalancerInternalListerner = "aws_internal_lb_listener"
	LoadBalancerListener          = "aws_lb_listener"

	LoadBalancerType        = "load_balancer_type"
	LoadBalancerTargetGroup = "aws_lb_target_group"

	LoadBalancerInternalTargetGroupAttachment = "aws_internal_lb_target_group_attachment"
	LoadBalancerTargetGroupAttachment         = "aws_lb_target_group_attachment"

	DefaultAction = "default_action"
	HealthCheck   = "health_check"

	Alias                 = "alias"
	Route53InternalRecord = "aws_internal_route53_record"
	Route53Record         = "aws_route53_record"
	Route53Zone           = "aws_route53_zone"
	Subnets               = "subnets"
	SubnetMapping         = "subnet_mapping"

	InternalTargetGroup80Attachment   = "aws_internal_tg_attachment_80_server"
	InternalTargetGroup443Attachment  = "aws_internal_tg_attachment_443_server"
	InternalTargetGroup6443Attachment = "aws_internal_tg_attachment_6443_server"
	InternalTargetGroup9345Attachment = "aws_internal_tg_attachment_9345_server"
	TargetGroup80Attachment           = "aws_tg_attachment_80_server"
	TargetGroup443Attachment          = "aws_tg_attachment_443_server"
	TargetGroup6443Attachment         = "aws_tg_attachment_6443_server"
	TargetGroup9345Attachment         = "aws_tg_attachment_9345_server"

	TargetGroupARN            = "target_group_arn"
	TargetGroupInternalPrefix = "aws_internal_tg_"
	TargetGroupPrefix         = "aws_tg_"
	TargetID                  = "target_id"
)
