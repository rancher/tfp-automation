package imported

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/imported"
	"github.com/rancher/tfp-automation/framework/set/resources/sanity/aws"
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
	token            = "token"
	sshKey           = "ssh_key"
	user             = "user"
)

// // SetImportedRKE2K3s is a function that will set the imported RKE2/K3s cluster configurations in the main.tf file.
func SetImportedRKE2K3s(rancherConfig *rancher.Config, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	clusterName string, newFile *hclwrite.File, rootBody *hclwrite.Body, file *os.File) (*os.File, error) {
	SetImportedCluster(rootBody, clusterName)

	rootBody.AppendNewline()
	serverOneName := clusterName + `_` + serverOne
	serverTwoName := clusterName + `_` + serverTwo
	serverThreeName := clusterName + `_` + serverThree
	instances := []string{serverOneName, serverTwoName, serverThreeName}

	for _, instance := range instances {
		aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	var nodeOnePublicDNS, nodeOnePrivateIP, nodeTwoPublicDNS, nodeThreePublicDNS string

	nodeOnePrivateIP = fmt.Sprintf("${%s.%s.private_ip}", defaults.AwsInstance, serverOneName)
	nodeOnePublicDNS = fmt.Sprintf("${%s.%s.public_dns}", defaults.AwsInstance, serverOneName)
	nodeTwoPublicDNS = fmt.Sprintf("${%s.%s.public_dns}", defaults.AwsInstance, serverTwoName)
	nodeThreePublicDNS = fmt.Sprintf("${%s.%s.public_dns}", defaults.AwsInstance, serverThreeName)

	imported.CreateRKE2K3SImportedCluster(rootBody, terraformConfig, nodeOnePublicDNS, nodeOnePrivateIP, nodeTwoPublicDNS, nodeThreePublicDNS, clusterName)

	rootBody.AppendNewline()

	importCommand := getImportCommand(clusterName)

	err := importNodes(rootBody, terraformConfig, nodeOnePublicDNS, "", importCommand[serverOneName], clusterName)
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
