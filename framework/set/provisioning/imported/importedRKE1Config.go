package imported

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

// // SetImportedRKE1 is a function that will set the imported RKE1 cluster configurations in the main.tf file.
func SetImportedRKE1(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File,
	rootBody *hclwrite.Body, file *os.File) (*hclwrite.File, *os.File, error) {
	SetImportedCluster(rootBody, terraformConfig.ResourcePrefix)

	rootBody.AppendNewline()

	createRKE1Cluster(rootBody, terraformConfig, terratestConfig)

	importCommand := getImportCommand(terraformConfig.ResourcePrefix)

	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne
	nodeOnePublicDNS := fmt.Sprintf("${%s.%s.public_dns}", defaults.AwsInstance, serverOneName)
	kubeConfig := fmt.Sprintf("${%s.%s.kube_config_yaml}", defaults.RKECluster, terraformConfig.ResourcePrefix)

	err := importNodes(rootBody, terraformConfig, nodeOnePublicDNS, kubeConfig, importCommand[serverOneName])
	if err != nil {
		return nil, nil, err
	}

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to write imported RKE1 configurations to main.tf file. Error: %v", err)
		return nil, nil, err
	}

	return newFile, file, nil
}

// createRKE1Cluster is a helper function that will create the RKE1 cluster.
func createRKE1Cluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) {
	serverOneName := terraformConfig.ResourcePrefix + `_` + serverOne
	serverTwoName := terraformConfig.ResourcePrefix + `_` + serverTwo
	serverThreeName := terraformConfig.ResourcePrefix + `_` + serverThree
	instances := []string{serverOneName, serverTwoName, serverThreeName}

	for _, instance := range instances {
		aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
		rootBody.AppendNewline()
	}

	rkeBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.RKECluster, terraformConfig.ResourcePrefix})
	rkeBlockBody := rkeBlock.Body()

	for _, instance := range instances {
		nodesBlock := rkeBlockBody.AppendNewBlock(defaults.Nodes, nil)
		nodesBlockBody := nodesBlock.Body()

		addressExpression := `"${` + defaults.AwsInstance + "." + instance + ".public_ip" + `}"`
		values := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(addressExpression)},
		}

		nodesBlockBody.SetAttributeRaw(address, values)
		nodesBlockBody.SetAttributeValue(user, cty.StringVal(terraformConfig.Standalone.OSUser))

		rolesExpression := fmt.Sprintf(`["controlplane", "etcd", "worker"]`)
		values = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(rolesExpression)},
		}

		nodesBlockBody.SetAttributeRaw(role, values)

		keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		nodesBlockBody.SetAttributeRaw(sshKey, keyPath)
	}

	rkeBlockBody.SetAttributeValue(enableCriDockerD, cty.BoolVal(true))

	var dependsOnServer string

	dependsOnServer = `[` + defaults.AwsInstance + `.` + serverOneName + `, ` + defaults.AwsInstance + `.` + serverTwoName + `, ` + defaults.AwsInstance + `.` + serverThreeName + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	rkeBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	rootBody.AppendNewline()
}
