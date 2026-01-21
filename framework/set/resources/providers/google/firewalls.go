package google

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	googleDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/zclconf/go-cty/cty"
)

const (
	all                  = "all"
	allow                = "allow"
	allowInternal        = "allow_internal"
	allowSSH             = "allow_ssh"
	allowTCP             = "allow_tcp"
	allowTCPLoadBalancer = "allow-tcp-load-balancer"
	backend              = "backend"
	balancingMode        = "balancing_mode"
	backendService       = "backend_service"
	defaultNetwork       = "default"
	group                = "group"
	healthChecks         = "health_checks"
	instances            = "instances"
	internalAccess       = "allow-internal"
	ipProtocol           = "ip_protocol"
	loadBalancerScheme   = "load_balancing_scheme"
	machineType          = "machine_type"
	namedPort            = "named_port"
	network              = "network"
	protocol             = "protocol"
	server1              = "server1"
	server2              = "server2"
	server3              = "server3"
	sshAccess            = "allow-ssh"
	sshKeys              = "ssh-keys"
	sourceRanges         = "source_ranges"
	tcp                  = "tcp"
	tcpHealthCheck       = "tcp_health_check"
	timeoutSecond        = "timeout_sec"
)

// CreateGoogleCloudFirewalls will set up the Google Cloud firewall rules.
func CreateGoogleCloudFirewalls(rootBody *hclwrite.Body) {
	firewallInternalBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeFirewall, allowInternal})
	firewallInternalBlockBody := firewallInternalBlock.Body()

	firewallInternalBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(internalAccess))
	firewallInternalBlockBody.SetAttributeValue(network, cty.StringVal(defaultNetwork))

	allowInternalBlock := firewallInternalBlockBody.AppendNewBlock(allow, nil)
	allowInternalBlockBody := allowInternalBlock.Body()

	allowInternalBlockBody.SetAttributeValue(protocol, cty.StringVal(all))

	internalSourceRangesList := []cty.Value{cty.StringVal("10.128.0.0/9")}
	internalSourceRangesCtyList := cty.ListVal(internalSourceRangesList)
	firewallInternalBlockBody.SetAttributeValue(sourceRanges, internalSourceRangesCtyList)

	rootBody.AppendNewline()

	firewallTCPBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeFirewall, allowTCP})
	firewallTCPBlockBody := firewallTCPBlock.Body()

	firewallTCPBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(allowTCPLoadBalancer))
	firewallTCPBlockBody.SetAttributeValue(network, cty.StringVal(defaultNetwork))

	allowBlock := firewallTCPBlockBody.AppendNewBlock(allow, nil)
	allowBlockBody := allowBlock.Body()

	allowBlockBody.SetAttributeValue(protocol, cty.StringVal(tcp))

	portsList := []cty.Value{cty.StringVal("80"), cty.StringVal("443"), cty.StringVal("6443"), cty.StringVal("9345")}
	portsCtyList := cty.ListVal(portsList)
	allowBlockBody.SetAttributeValue(general.Ports, portsCtyList)

	sourceRangesList := []cty.Value{cty.StringVal("0.0.0.0/0")}
	sourceRangesCtyList := cty.ListVal(sourceRangesList)
	firewallTCPBlockBody.SetAttributeValue(sourceRanges, sourceRangesCtyList)

	rootBody.AppendNewline()

	firewallSSHBlock := rootBody.AppendNewBlock(general.Resource, []string{googleDefaults.GoogleComputeFirewall, allowSSH})
	firewallSSHBlockBody := firewallSSHBlock.Body()

	firewallSSHBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(sshAccess))
	firewallSSHBlockBody.SetAttributeValue(network, cty.StringVal(defaultNetwork))

	allowSSHBlock := firewallSSHBlockBody.AppendNewBlock(allow, nil)
	allowSSHBlockBody := allowSSHBlock.Body()

	allowSSHBlockBody.SetAttributeValue(protocol, cty.StringVal(tcp))

	sshPortsList := []cty.Value{cty.StringVal("22")}
	sshPortsCtyList := cty.ListVal(sshPortsList)
	allowSSHBlockBody.SetAttributeValue(general.Ports, sshPortsCtyList)

	sshSourceRangesList := []cty.Value{cty.StringVal("0.0.0.0/0")}
	sshSourceRangesCtyList := cty.ListVal(sshSourceRangesList)
	firewallSSHBlockBody.SetAttributeValue(sourceRanges, sshSourceRangesCtyList)
}
