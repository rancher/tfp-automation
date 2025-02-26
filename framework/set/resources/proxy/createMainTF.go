package proxy

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/proxy/squid"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
	"github.com/sirupsen/logrus"
)

const (
	rke2Bastion     = "rke2_bastion"
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"

	rke2ServerOnePublicDNS   = "rke2_server1_public_dns"
	rke2BastionPublicDNS     = "rke2_bastion_public_dns"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPublicDNS   = "rke2_server2_public_dns"
	rke2ServerThreePublicDNS = "rke2_server3_public_dns"

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

	instances := []string{rke2Bastion, rke2ServerOne, rke2ServerTwo, rke2ServerThree}

	logrus.Infof("Creating AWS resources...")
	file, err := aws.CreateAWSResources(file, newFile, tfBlockBody, rootBody, terraformConfig, terratest, instances)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	rke2BastionPublicDNS := terraform.Output(t, terraformOptions, rke2BastionPublicDNS)
	rke2ServerOnePublicDNS := terraform.Output(t, terraformOptions, rke2ServerOnePublicDNS)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPublicDNS := terraform.Output(t, terraformOptions, rke2ServerTwoPublicDNS)
	rke2ServerThreePublicDNS := terraform.Output(t, terraformOptions, rke2ServerThreePublicDNS)

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating squid proxy...")
	file, err = squid.CreateSquidProxy(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOnePublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPublicDNS, rke2ServerThreePublicDNS)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = sanity.OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateProxiedRancher(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicDNS, rke2BastionPublicDNS)
	if err != nil {
		return "", "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	return rke2BastionPublicDNS, rke2ServerOnePublicDNS, nil
}
