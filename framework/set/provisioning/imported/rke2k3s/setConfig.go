package rke2k3s

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	awsDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/sleep"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported"
	resources "github.com/rancher/tfp-automation/framework/set/resources/imported"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
)

const (
	address          = "address"
	addServer        = "add_server_"
	importCluster    = "import_cluster"
	copyScript       = "copy_script"
	enableCriDockerD = "enable_cri_dockerd"
	role             = "role"
	serverOne        = "server1"
	serverTwo        = "server2"
	serverThree      = "server3"
	sshKey           = "ssh_key"
	user             = "user"
	windows          = "windows"
)

// // SetImportedRKE2K3s is a function that will set the imported RKE2/K3s cluster configurations in the main.tf file.
func SetImportedRKE2K3s(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	imported.SetImportedCluster(rootBody, terraformConfig.ResourcePrefix)
	rootBody.AppendNewline()

	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne
	windowsNodeName := terraformConfig.ResourcePrefix + `-` + windows

	nodeOnePublicIP, nodeOnePrivateIP, nodeTwoPublicIP, nodeThreePublicIP := getProviderIPAddresses(terraformConfig, terratestConfig, rootBody, serverOneName)

	token := namegen.AppendRandomString(general.Import)

	resources.CreateRKE2K3SImportedCluster(rootBody, terraformConfig, terratestConfig, nodeOnePublicIP, nodeOnePrivateIP, nodeTwoPublicIP, nodeThreePublicIP, token)
	rootBody.AppendNewline()

	if strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) {
		aws.CreateWindowsAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
		rootBody.AppendNewline()

		windowsNodePublicDNS := fmt.Sprintf("${%s.%s.public_dns}", awsDefaults.AwsInstance, windowsNodeName)
		resources.AddWindowsNodeToImportedCluster(rootBody, terraformConfig, terratestConfig, nodeOnePrivateIP, windowsNodePublicDNS, token)

		// Add the sleep command to wait for the Windows node to be ready
		rootBody.AppendNewline()
		dependsOnValue := fmt.Sprintf("[" + general.NullResource + ".add_windows_node" + "]")

		sleep.SetTimeSleep(rootBody, terraformConfig, "10s", dependsOnValue, "import_wins")
		rootBody.AppendNewline()
	}

	importCommand := imported.GetImportCommand(terraformConfig.ResourcePrefix)

	err := imported.ImportNodes(rootBody, terraformConfig, terratestConfig, nodeOnePublicIP, "", importCommand[serverOneName])
	if err != nil {
		return nil, nil, err
	}

	return newFile, file, nil
}
