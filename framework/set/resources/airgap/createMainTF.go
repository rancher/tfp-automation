package airgap

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	shepherdConfig "github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/cleanup"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	registry "github.com/rancher/tfp-automation/framework/set/resources/registries/createRegistry"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
)

const (
	rancherRegistry = "registry"
	bastion         = "bastion"
	serverOne       = "server1"
	serverTwo       = "server2"
	serverThree     = "server3"

	nonAuthRegistry = "non_auth_registry"

	registryPublicDNS    = "registry_public_dns"
	bastionPublicDNS     = "bastion_public_dns"
	serverOnePrivateIP   = "server1_private_ip"
	serverTwoPrivateIP   = "server2_private_ip"
	serverThreePrivateIP = "server3_private_ip"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating an Airgapped-Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, rancherConfig *shepherdConfig.Config,
	terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) (string, string, error) {
	var file *os.File
	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{bastion, rancherRegistry}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating resources. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	registryPublicDNS := terraform.Output(t, terraformOptions, registryPublicDNS)
	bastionPublicDNS := terraform.Output(t, terraformOptions, bastionPublicDNS)
	serverOnePrivateIP := terraform.Output(t, terraformOptions, serverOnePrivateIP)
	serverTwoPrivateIP := terraform.Output(t, terraformOptions, serverTwoPrivateIP)
	serverThreePrivateIP := terraform.Output(t, terraformOptions, serverThreePrivateIP)

	logrus.Infof("Creating registry...")
	file = sanity.OpenFile(file, keyPath)
	file, err = registry.CreateNonAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, registryPublicDNS, nonAuthRegistry)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating registry. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateAirgapRKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, registryPublicDNS, serverOnePrivateIP, serverTwoPrivateIP, serverThreePrivateIP)
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
	file, err = rancher.CreateAirgapRancher(file, newFile, rootBody, terraformConfig, terratestConfig, bastionPublicDNS, registryPublicDNS)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating Rancher server. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	return registryPublicDNS, bastionPublicDNS, nil
}
