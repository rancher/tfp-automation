package imported

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/vsphere"
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

	var nodeOnePublicDNS string
	if terraformConfig.Provider == defaults.Aws {
		nodeOnePublicDNS = fmt.Sprintf("${%s.%s.public_dns}", defaults.AwsInstance, serverOneName)
	} else if terraformConfig.Provider == defaults.Vsphere {
		nodeOnePublicDNS = fmt.Sprintf("${%s.%s.default_ip_address}", defaults.VsphereVirtualMachine, serverOneName)
	}

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

	if terraformConfig.Provider == defaults.Vsphere {
		dataCenterExpression := fmt.Sprintf(defaults.Data + `.` + defaults.VsphereDatacenter + `.` + defaults.VsphereDatacenter + `.id`)
		dataCenterValue := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dataCenterExpression)},
		}

		vsphere.CreateVsphereDatacenter(rootBody, terraformConfig)
		rootBody.AppendNewline()

		vsphere.CreateVsphereComputeCluster(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()

		vsphere.CreateVsphereNetwork(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()

		vsphere.CreateVsphereDatastore(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()

		vsphere.CreateVsphereVirtualMachineTemplate(rootBody, terraformConfig, dataCenterValue)
		rootBody.AppendNewline()
	}

	for _, instance := range instances {
		if terraformConfig.Provider == defaults.Aws {
			aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()
		} else if terraformConfig.Provider == defaults.Vsphere {
			vsphere.CreateVsphereVirtualMachine(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()
		}
	}

	rkeBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.RKECluster, terraformConfig.ResourcePrefix})
	rkeBlockBody := rkeBlock.Body()

	for _, instance := range instances {
		nodesBlock := rkeBlockBody.AppendNewBlock(defaults.Nodes, nil)
		nodesBlockBody := nodesBlock.Body()

		var addressExpression string
		if terraformConfig.Provider == defaults.Aws {
			addressExpression = `"${` + defaults.AwsInstance + "." + instance + ".public_ip" + `}"`
		} else if terraformConfig.Provider == defaults.Vsphere {
			addressExpression = `"${` + defaults.VsphereVirtualMachine + "." + instance + ".default_ip_address" + `}"`
		}

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
	if terraformConfig.Provider == defaults.Aws {
		dependsOnServer = `[` + defaults.AwsInstance + `.` + serverOneName + `, ` + defaults.AwsInstance + `.` + serverTwoName + `, ` + defaults.AwsInstance + `.` + serverThreeName + `]`
	} else if terraformConfig.Provider == defaults.Vsphere {
		dependsOnServer = `[` + defaults.VsphereVirtualMachine + `.` + serverOneName + `, ` + defaults.VsphereVirtualMachine + `.` + serverTwoName + `, ` + defaults.VsphereVirtualMachine + `.` + serverThreeName + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	rkeBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	rootBody.AppendNewline()
}
