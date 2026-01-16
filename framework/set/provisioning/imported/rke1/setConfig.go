package rke1

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	awsDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	vsphereDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/defaults/rke"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported"
)

const (
	address          = "address"
	addServer        = "add_server_"
	enableCriDockerD = "enable_cri_dockerd"
	importCluster    = "import_cluster"
	copyScript       = "copy_script"
	role             = "role"
	serverOne        = "server1"
	serverTwo        = "server2"
	serverThree      = "server3"
	sshKey           = "ssh_key"
	user             = "user"
	windows          = "windows"
)

// // SetImportedRKE1 is a function that will set the imported RKE1 cluster configurations in the main.tf file.
func SetImportedRKE1(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	imported.SetImportedCluster(rootBody, terraformConfig.ResourcePrefix)

	rootBody.AppendNewline()

	createRKE1Cluster(rootBody, terraformConfig, terratestConfig)

	importCommand := imported.GetImportCommand(terraformConfig.ResourcePrefix)

	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne

	var nodeOnePublicDNS string
	switch terraformConfig.Provider {
	case awsDefaults.Aws:
		nodeOnePublicDNS = fmt.Sprintf("${%s.%s.public_dns}", awsDefaults.AwsInstance, serverOneName)
	case vsphereDefaults.Vsphere:
		nodeOnePublicDNS = fmt.Sprintf("${%s.%s.default_ip_address}", vsphereDefaults.VsphereVirtualMachine, serverOneName)
	}

	kubeConfig := fmt.Sprintf("${%s.%s.kube_config_yaml}", rke.RKECluster, terraformConfig.ResourcePrefix)

	err := imported.ImportNodes(rootBody, terraformConfig, terratestConfig, nodeOnePublicDNS, kubeConfig, importCommand[serverOneName])
	if err != nil {
		return nil, nil, err
	}

	return newFile, file, nil
}
