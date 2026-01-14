package proxy

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	tunnel "github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/squid"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
)

const (
	nodeBalancerHostname = "linode_node_balancer_hostname"

	bastion     = "bastion"
	serverOne   = "server1"
	serverTwo   = "server2"
	serverThree = "server3"

	bastionPublicDNS     = "bastion_public_dns"
	bastionPrivateIP     = "bastion_private_ip"
	serverOnePrivateIP   = "server1_private_ip"
	serverTwoPrivateIP   = "server2_private_ip"
	serverThreePrivateIP = "server3_private_ip"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating a Rancher server behind a proxy.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, rancherConfig *shepherdConfig.Config,
	terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (string, string, error) {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	var err error

	instances := []string{bastion}

	providerTunnel := tunnel.TunnelToProvider(terraformConfig.Provider)
	file, err = providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating resources. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	bastionPublicDNS := terraform.Output(t, terraformOptions, bastionPublicDNS)
	bastionPrivateIP := terraform.Output(t, terraformOptions, bastionPrivateIP)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverTwoPrivateIP := terraform.Output(t, terraformOptions, serverTwoPrivateIP)
	serverThreePrivateIP := terraform.Output(t, terraformOptions, serverThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating squid proxy...")
	file, err = squid.CreateSquidProxy(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating squid proxy. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, bastionPrivateIP, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating RKE2 cluster. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateProxiedRancher(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, bastionPrivateIP)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating Rancher server. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	terraformConfig.Proxy.ProxyBastion = bastionPublicDNS

	return bastionPublicDNS, bastionPrivateIP, nil
}
