package rke1

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	awsDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	vsphereDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/defaults/rke"
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

	if terraformConfig.Provider == vsphereDefaults.Vsphere {
		dataCenterExpression := fmt.Sprintf(general.Data + `.` + vsphereDefaults.VsphereDatacenter + `.` + vsphereDefaults.VsphereDatacenter + `.id`)
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
		case awsDefaults.Aws:
			aws.CreateAWSInstances(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()
		case vsphereDefaults.Vsphere:
			vsphere.CreateVsphereVirtualMachine(rootBody, terraformConfig, terratestConfig, instance)
			rootBody.AppendNewline()
		}
	}

	rkeBlock := rootBody.AppendNewBlock(general.Resource, []string{rke.RKECluster, terraformConfig.ResourcePrefix})
	rkeBlockBody := rkeBlock.Body()

	for _, instance := range instances {
		nodesBlock := rkeBlockBody.AppendNewBlock(awsDefaults.Nodes, nil)
		nodesBlockBody := nodesBlock.Body()

		var addressExpression string
		switch terraformConfig.Provider {
		case awsDefaults.Aws:
			addressExpression = `"${` + awsDefaults.AwsInstance + "." + instance + ".public_ip" + `}"`
		case vsphereDefaults.Vsphere:
			addressExpression = `"${` + vsphereDefaults.VsphereVirtualMachine + "." + instance + ".default_ip_address" + `}"`
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

		keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		nodesBlockBody.SetAttributeRaw(sshKey, keyPath)
	}

	rkeBlockBody.SetAttributeValue(enableCriDockerD, cty.BoolVal(true))

	var dependsOnServer string
	switch terraformConfig.Provider {
	case awsDefaults.Aws:
		dependsOnServer = `[` + awsDefaults.AwsInstance + `.` + serverOneName + `, ` + awsDefaults.AwsInstance + `.` + serverTwoName + `, ` + awsDefaults.AwsInstance + `.` + serverThreeName + `]`
	case vsphereDefaults.Vsphere:
		dependsOnServer = `[` + vsphereDefaults.VsphereVirtualMachine + `.` + serverOneName + `, ` + vsphereDefaults.VsphereVirtualMachine + `.` + serverTwoName + `, ` + vsphereDefaults.VsphereVirtualMachine + `.` + serverThreeName + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	rkeBlockBody.SetAttributeRaw(general.DependsOn, server)

	rootBody.AppendNewline()
}
