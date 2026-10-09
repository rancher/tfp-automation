package ipv6

import (
	"context"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/scripts"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/k3s"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/rancher"
	"github.com/sirupsen/logrus"
)

const (
	bastion     = "bastion"
	serverOne   = "server1"
	serverTwo   = "server2"
	serverThree = "server3"

	bastionPublicIP      = "bastion_public_ip"
	serverOnePrivateIP   = "server1_private_ip"
	serverOnePublicIP    = "server1_public_ip"
	serverTwoPrivateIP   = "server2_private_ip"
	serverTwoPublicIP    = "server2_public_ip"
	serverThreePrivateIP = "server3_private_ip"
	serverThreePublicIP  = "server3_public_ip"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating an Airgapped-Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, rancherConfig *shepherdConfig.Config,
	terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (string, error) {
	var file *os.File

	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{bastion}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateIPv6(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", err
	}

	_, err = scripts.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating resources. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", err
	}

	bastionPublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverOnePrivateIP)
	serverOnePublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverOnePublicIP)
	serverTwoPrivateIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverTwoPrivateIP)
	serverTwoPublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverTwoPublicIP)
	serverThreePrivateIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverThreePrivateIP)
	serverThreePublicIP := terraform.OutputContext(t, context.Background(), terraformOptions, serverThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	if terraformConfig.LocalCluster == "k3s" {
		logrus.Infof("Creating K3S cluster...")
		file, err = k3s.CreateIPv6K3SCluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP, serverOnePublicIP, serverTwoPublicIP, serverThreePublicIP,
			serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
		if err != nil {
			return "", err
		}
	} else if terraformConfig.LocalCluster == "rke2" {
		logrus.Infof("Creating RKE2 cluster...")
		file, err = rke2.CreateIPv6RKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP, serverOnePublicIP, serverTwoPublicIP, serverThreePublicIP,
			serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
		if err != nil {
			return "", err
		}
	}

	_, err = scripts.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating local cluster. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", err
	}

	logrus.Infof("Creating Rancher server...")
	file = sanity.OpenFile(file, keyPath)
	file, err = rancher.CreateRancher(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP)
	if err != nil {
		return "", err
	}

	_, err = scripts.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating Rancher server. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", err
	}

	return bastionPublicIP, nil
}
