package rke2

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"
	token           = "token"
)

// CreateRKE2Cluster is a helper function that will create the RKE2 cluster.
func CreateRKE2Cluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	serverScriptPath := filepath.Join(userDir, "go/", terratestConfig.PathToRepo, "/framework/set/resources/rke2/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, "go/", terratestConfig.PathToRepo, "/framework/set/resources/rke2/add-servers.sh")

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return nil, err
	}

	newServersScriptContent, err := os.ReadFile(newServersScriptPath)
	if err != nil {
		return nil, err
	}

	rke2Token := namegen.AppendRandomString(token)

	createRKE2Server(rootBody, terraformConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2Token, serverOneScriptContent)
	addRKE2ServerNodes(rootBody, terraformConfig, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP, rke2Token, newServersScriptContent)

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// SSHNullResource is a helper function that will create the null_resource to SSH into the instance.
func SSHNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, instance, host string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, host})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(defaults.Host, cty.StringVal(instance))
	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))

	if terraformConfig.Provider == defaults.Aws {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

		keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	} else if terraformConfig.Provider == defaults.Linode {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(linode.RootUser))
		connectionBlockBody.SetAttributeValue(defaults.Password, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	} else if terraformConfig.Provider == defaults.Harvester {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.HarvesterConfig.SSHUser))

		keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	} else if terraformConfig.Provider == defaults.Vsphere {
		connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.VsphereConfig.VsphereUser))

		keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)
	}

	rootBody.AppendNewline()

	return nullResourceBlockBody, provisionerBlockBody
}

// createRKE2Server is a helper function that will create the RKE2 server.
func createRKE2Server(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP,
	rke2Token string, script []byte) {
	_, provisionerBlockBody := SSHNullResource(rootBody, terraformConfig, rke2ServerOnePublicIP, rke2ServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " + rke2Token + " " + terraformConfig.CNI + " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
		cty.StringVal(command),
	}))
}

// addRKE2ServerNodes is a helper function that will add additional RKE2 server nodes to the initial RKE2 server.
func addRKE2ServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP,
	rke2ServerThreePublicIP, rke2Token string, script []byte) {
	instances := []string{rke2ServerTwoPublicIP, rke2ServerThreePublicIP}
	hosts := []string{rke2ServerTwo, rke2ServerThree}

	for i, instance := range instances {
		host := hosts[i]
		nullResourceBlockBody, provisionerBlockBody := SSHNullResource(rootBody, terraformConfig, instance, host)

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " +
			instance + " " + rke2Token + " " + terraformConfig.CNI + " || true'"

		provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("printf '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
			cty.StringVal(command),
		}))

		dependsOnServer := `[` + defaults.NullResource + `.` + rke2ServerOne + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
	}
}
