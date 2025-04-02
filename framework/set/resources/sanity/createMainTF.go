package sanity

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/configs"
	"github.com/rancher/tfp-automation/defaults/providers"
	tunnel "github.com/rancher/tfp-automation/framework/set/resources/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/rancher"
	"github.com/sirupsen/logrus"
)

const (
	nodeBalancerHostname    = "linode_node_balancer_hostname"
	rke2ServerOne           = "rke2_server1"
	rke2ServerTwo           = "rke2_server2"
	rke2ServerThree         = "rke2_server3"
	rke2ServerOnePublicIP   = "rke2_server1_public_ip"
	rke2ServerOnePrivateIP  = "rke2_server1_private_ip"
	rke2ServerTwoPublicIP   = "rke2_server2_public_ip"
	rke2ServerThreePublicIP = "rke2_server3_public_ip"
	terraformConst          = "terraform"
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

	var err error
	var linodeNodeBalancerHostname string

	instances := []string{rke2ServerOne, rke2ServerTwo, rke2ServerThree}

	providerTunnel := tunnel.TunnelToProvider(terraformConfig)
	file, err = providerTunnel.CreateNonAirgap(file, newFile, tfBlockBody, rootBody, terraformConfig, terratestConfig, instances)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	if terraformConfig.NodeProvider == providers.Linode {
		linodeNodeBalancerHostname = terraform.Output(t, terraformOptions, nodeBalancerHostname)
	}

	rke2ServerOnePublicIP := terraform.Output(t, terraformOptions, rke2ServerOnePublicIP)
	rke2ServerOnePrivateIP := terraform.Output(t, terraformOptions, rke2ServerOnePrivateIP)
	rke2ServerTwoPublicIP := terraform.Output(t, terraformOptions, rke2ServerTwoPublicIP)
	rke2ServerThreePublicIP := terraform.Output(t, terraformOptions, rke2ServerThreePublicIP)

	file = OpenFile(file, keyPath)
	logrus.Infof("Creating RKE2 cluster...")
	file, err = rke2.CreateRKE2Cluster(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	file = OpenFile(file, keyPath)
	logrus.Infof("Creating Rancher server...")
	file, err = rancher.CreateRancher(file, newFile, rootBody, terraformConfig, rke2ServerOnePublicIP, linodeNodeBalancerHostname)
	if err != nil {
		return "", err
	}

	terraform.InitAndApply(t, terraformOptions)

	return rke2ServerOnePublicIP, nil
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
