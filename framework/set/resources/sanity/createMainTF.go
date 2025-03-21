package sanity

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/rancher"
	"github.com/sirupsen/logrus"
)

const (
	rke2ServerOne            = "rke2_server1"
	rke2ServerTwo            = "rke2_server2"
	rke2ServerThree          = "rke2_server3"
	rke2ServerOnePublicDNS   = "rke2_server1_public_dns"
	rke2ServerOnePrivateIP   = "rke2_server1_private_ip"
	rke2ServerTwoPublicDNS   = "rke2_server2_public_dns"
	rke2ServerThreePublicDNS = "rke2_server3_public_dns"
	terraformConst           = "terraform"
)

// CreateMainTF is a helper function that will create the main.tf file for creating a Rancher server.
func CreateMainTF(t *testing.T, terraformOptions *terraform.Options, keyPath string, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig) (string, error) {
	var file *os.File
	file = OpenFile(file, keyPath)
	defer file.Close()

	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()

	tfBlock := rootBody.AppendNewBlock(terraformConst, nil)
	tfBlockBody := tfBlock.Body()

	instances := []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}

	logrus.Infof("Creating AWS resources...")
	file, err := aws.CreateAWSResources(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	rke2ServerOnePublicDNS := terraform.Output(t, terraformOptions, rke2ServerOnePublicDNS)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPublicDNS := terraform.Output(t, terraformOptions, rke2ServerTwoPublicDNS)
	rke2ServerThreePublicDNS := terraform.Output(t, terraformOptions, rke2ServerThreePublicDNS)

	file = OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPublicDNS, rke2ServerThreePublicDNS)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateRancher(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicDNS)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	return rke2ServerOnePublicDNS, nil
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
