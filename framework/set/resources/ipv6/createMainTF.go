package airgap

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/ipv6/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/sirupsen/logrus"
)

const (
	rancherRegistry = "registry"
	rke2Bastion     = "rke2_bastion"
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"

	rke2BastionPublicDNS     = "rke2_bastion_public_dns"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPrivateIP   = "rke2_server2_private_ip"
	rke2ServerThreePrivateIP = "rke2_server3_private_ip"

	terraformConst = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating an Airgapped-Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig) (string, error) {
	var file *os.File

	file = sanity.OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{rke2Bastion}

	providerTunnel := providers.TunnelToProvider(terraformConfig.Provider)
	file, err := providerTunnel.CreateAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	rke2BastionPublicDNS := terraform.Output(t, terraformOptions, rke2BastionPublicDNS)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPrivateIP := terraform.Output(t, terraformOptions, rke2ServerTwoPrivateIP)
	rke2ServerThreePrivateIP := terraform.Output(t, terraformOptions, rke2ServerThreePrivateIP)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateIPv6Cluster(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	logrus.Infof("Creating Rancher server...")
	file = sanity.OpenFile(file, keyPath)
	file, err = rancher.CreateIPv6Rancher(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	return rke2BastionPublicDNS, nil
}
