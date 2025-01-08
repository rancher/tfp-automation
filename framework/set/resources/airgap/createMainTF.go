package airgap

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap/rancher"
	"github.com/rancher/tfp-automation/framework/set/resources/airgap/rke2"
	"github.com/sirupsen/logrus"
)

const (
	rke2ServerOne            = "rke2_server1"
	rke2ServerTwo            = "rke2_server2"
	rke2ServerThree          = "rke2_server3"
	rke2BastionPublicDNS     = "rke2_bastion_public_dns"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPrivateIP   = "rke2_server2_private_ip"
	rke2ServerThreePrivateIP = "rke2_server3_private_ip"
	terraformConst           = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating an Airgapped-Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, terraformConfig *config.TerraformConfig,
	terratest *config.TerratestConfig) error {
	var file *os.File
	file = OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	file, err := aws.CreateAWSResources(file, newFile, tfBlockBody, rootBody, terraformConfig, terratest)
	if err != nil {
		return err
	}

	terraform.InitAndApply(t, terraformOptions)

	rke2BastionPublicDNS := terraform.Output(t, terraformOptions, rke2BastionPublicDNS)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPrivateIP := terraform.Output(t, terraformOptions, rke2ServerTwoPrivateIP)
	rke2ServerThreePrivateIP := terraform.Output(t, terraformOptions, rke2ServerThreePrivateIP)

	file = OpenFile(file, keyPath)
	file, err = rke2.CreateAirgapRKE2Cluster(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP)
	if err != nil {
		return err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = OpenFile(file, keyPath)
	file, err = rancher.CreateAirgapRancher(file, newFile, rootBody, terraformConfig, rke2BastionPublicDNS)
	if err != nil {
		return err
	}

	terraform.InitAndApply(t, terraformOptions)

	return nil
}

// OpenFile is a helper function that will open the main.tf file.
func OpenFile(file *os.File, keyPath string) *os.File {
	file, err := os.Create(keyPath + configs.MainTF)
	if err != nil {
		logrus.Infof("Failed to reset/overwrite main.tf file. Error: %v", err)
		return nil
	}

	return file
}
