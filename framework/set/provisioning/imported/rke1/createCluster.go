package rke1

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/vsphere"
	"github.com/zclconf/go-cty/cty"
)

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
		switch terraformConfig.Provider {
		case defaults.Aws:
			aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()
		case defaults.Vsphere:
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
		switch terraformConfig.Provider {
		case defaults.Aws:
			addressExpression = `"${` + defaults.AwsInstance + "." + instance + ".public_ip" + `}"`
		case defaults.Vsphere:
			addressExpression = `"${` + defaults.VsphereVirtualMachine + "." + instance + ".default_ip_address" + `}"`
		}

		values := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(addressExpression)},
		}

		nodesBlockBody.SetAttributeRaw(address, values)
		nodesBlockBody.SetAttributeValue(user, cty.StringVal(terraformConfig.Standalone.OSUser))

		rolesExpression := `["controlplane", "etcd", "worker"]`
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
	switch terraformConfig.Provider {
	case defaults.Aws:
		dependsOnServer = `[` + defaults.AwsInstance + `.` + serverOneName + `, ` + defaults.AwsInstance + `.` + serverTwoName + `, ` + defaults.AwsInstance + `.` + serverThreeName + `]`
	case defaults.Vsphere:
		dependsOnServer = `[` + defaults.VsphereVirtualMachine + `.` + serverOneName + `, ` + defaults.VsphereVirtualMachine + `.` + serverTwoName + `, ` + defaults.VsphereVirtualMachine + `.` + serverThreeName + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	rkeBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	rootBody.AppendNewline()
}
