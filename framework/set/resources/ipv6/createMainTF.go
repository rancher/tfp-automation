package ipv6

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
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

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating resources. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", err
	}

	bastionPublicIP := terraform.Output(t, terraformOptions, bastionPublicIP)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverOnePublicIP := terraform.Output(t, terraformOptions, serverOnePublicIP)
	serverTwoPrivateIP := terraform.Output(t, terraformOptions, serverTwoPrivateIP)
	serverTwoPublicIP := terraform.Output(t, terraformOptions, serverTwoPublicIP)
	serverThreePrivateIP := terraform.Output(t, terraformOptions, serverThreePrivateIP)
	serverThreePublicIP := terraform.Output(t, terraformOptions, serverThreePublicIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateIPv6RKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP, serverOnePublicIP, serverTwoPublicIP, serverThreePublicIP,
		serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
	if err != nil {
		return "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating RKE2 cluster. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", err
	}

	logrus.Infof("Creating Rancher server...")
	file = sanity.OpenFile(file, keyPath)
	file, err = rancher.CreateRancher(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicIP)
	if err != nil {
		return "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating Rancher server. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", err
	}

	return bastionPublicIP, nil
}
