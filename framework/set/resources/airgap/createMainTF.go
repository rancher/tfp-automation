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
	rke2Bastion     = "rke2_bastion"
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"

	registryPublicDNS        = "registry_public_dns"
	rke2BastionPublicDNS     = "rke2_bastion_public_dns"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPrivateIP   = "rke2_server2_private_ip"
	rke2ServerThreePrivateIP = "rke2_server3_private_ip"

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

	instances := []string{rke2Bastion, rancherRegistry}

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
	rke2BastionPublicDNS := terraform.Output(t, terraformOptions, rke2BastionPublicDNS)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPrivateIP := terraform.Output(t, terraformOptions, rke2ServerTwoPrivateIP)
	rke2ServerThreePrivateIP := terraform.Output(t, terraformOptions, rke2ServerThreePrivateIP)

	logrus.Infof("Creating registry...")
	file = sanity.OpenFile(file, keyPath)
	file, err = registry.CreateAuthenticatedRegistry(file, newFile, rootBody, terraformConfig, terratestConfig, registryPublicDNS)
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
	file, err = rke2.CreateAirgapRKE2Cluster(file, newFile, rootBody, terraformConfig, terratestConfig, rke2BastionPublicDNS, registryPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP)
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
	file, err = rancher.CreateAirgapRancher(file, newFile, rootBody, terraformConfig, terratestConfig, rke2BastionPublicDNS, registryPublicDNS)
	if err != nil {
		return "", "", err
	}

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	if err != nil && *rancherConfig.Cleanup {
		logrus.Infof("Error while creating Rancher server. Cleaning up...")
		cleanup.Cleanup(t, terraformOptions, keyPath)
		return "", "", err
	}

	return registryPublicDNS, rke2BastionPublicDNS, nil
}
