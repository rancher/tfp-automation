package proxy

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	tunnel "github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/squid"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
)

const (
	nodeBalancerHostname = "linode_node_balancer_hostname"

	rke2Bastion     = "rke2_bastion"
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"

	rke2BastionPublicDNS     = "rke2_bastion_public_dns"
	rke2BastionPrivateIP     = "rke2_bastion_private_ip"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPrivateIP   = "rke2_server2_private_ip"
	rke2ServerThreePrivateIP = "rke2_server3_private_ip"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating a Rancher server behind a proxy.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, terraformConfig *config.TerraformConfig,
	terratest *config.TerratestConfig) (string, string, error) {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	var err error
	var linodeNodeBalancerHostname string

	instances := []string{rke2Bastion}

	providerTunnel := tunnel.TunnelToProvider(terraformConfig.Provider)
	file, err = providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratest, instances)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	if terraformConfig.Provider == providers.Linode {
		linodeNodeBalancerHostname = terraform.Output(t, terraformOptions, nodeBalancerHostname)
	}

	rke2BastionPublicDNS := terraform.Output(t, terraformOptions, rke2BastionPublicDNS)
	rke2BastionPrivateIP := terraform.Output(t, terraformOptions, rke2BastionPrivateIP)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPrivateIP := terraform.Output(t, terraformOptions, rke2ServerTwoPrivateIP)
	rke2ServerThreePrivateIP := terraform.Output(t, terraformOptions, rke2ServerThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating squid proxy...")
	file, err = squid.CreateSquidProxy(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS, rke2BastionPrivateIP, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateProxiedRancher(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS, rke2BastionPrivateIP, linodeNodeBalancerHostname)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	terraformConfig.Proxy.ProxyBastion = rke2BastionPublicDNS

	return rke2BastionPublicDNS, rke2BastionPrivateIP, nil
}
