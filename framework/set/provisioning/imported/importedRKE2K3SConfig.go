package imported

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/custom/sleep"
	"github.com/rancher/tfp-automation/framework/set/resources/imported"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
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
	rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	SetImportedCluster(rootBody, terraformConfig.ResourcePrefix)

	rootBody.AppendNewline()

	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne
	serverTwoName := terraformConfig.ResourcePrefix + `_` + serverTwo
	serverThreeName := terraformConfig.ResourcePrefix + `_` + serverThree
	windowsNodeName := terraformConfig.ResourcePrefix + `-` + windows

	instances := []string{serverOneName, serverTwoName, serverThreeName}

	for _, instance := range instances {
		aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	var nodeOnePublicIP, nodeOnePrivateIP, nodeTwoPublicIP, nodeThreePublicIP string

	nodeOnePrivateIP = fmt.Sprintf("${%s.%s.private_ip}", defaults.AwsInstance, serverOneName)
	nodeOnePublicIP = fmt.Sprintf("${%s.%s.public_ip}", defaults.AwsInstance, serverOneName)
	nodeTwoPublicIP = fmt.Sprintf("${%s.%s.public_ip}", defaults.AwsInstance, serverTwoName)
	nodeThreePublicIP = fmt.Sprintf("${%s.%s.public_ip}", defaults.AwsInstance, serverThreeName)

	token := namegen.AppendRandomString(defaults.Import)

	imported.CreateRKE2K3SImportedCluster(rootBody, terraformConfig, nodeOnePublicIP, nodeOnePrivateIP, nodeTwoPublicIP, nodeThreePublicIP, token)

	rootBody.AppendNewline()

	if terraformConfig.Module == modules.ImportEC2RKE2Windows {
		aws.CreateWindowsAWSInstances(rootBody, terraformConfig, terratestConfig, terraformConfig.ResourcePrefix)
		rootBody.AppendNewline()

		windowsNodePublicDNS := fmt.Sprintf("${%s.%s.public_dns}", defaults.AwsInstance, windowsNodeName)
		imported.AddWindowsNodeToImportedCluster(rootBody, terraformConfig, terratestConfig, nodeOnePrivateIP, windowsNodePublicDNS, token)

		// Add the sleep command to wait for the Windows node to be ready
		rootBody.AppendNewline()
		dependsOnValue := fmt.Sprintf("[" + defaults.NullResource + ".add_windows_node" + "]")

		sleep.SetTimeSleep(rootBody, terraformConfig, "10s", dependsOnValue)
		rootBody.AppendNewline()
	}

	importCommand := getImportCommand(terraformConfig.ResourcePrefix)

	err := importNodes(rootBody, terraformConfig, nodeOnePublicIP, "", importCommand[serverOneName])
	if err != nil {
		return nil, err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write imported RKE2/K3s configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// getImportCommand is a helper function that will return the import command for the cluster
func getImportCommand(clusterName string) map[string]string {
	command := make(map[string]string)
	importCommand := fmt.Sprintf("${%s.%s.%s[0].%s}", defaults.Cluster, clusterName, defaults.ClusterRegistrationToken, defaults.InsecureCommand)

	serverOneName := clusterName + `_` + serverOne
	command[serverOneName] = importCommand

	return command
}
