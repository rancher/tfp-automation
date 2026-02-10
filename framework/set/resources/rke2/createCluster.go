package rke2

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/azure"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/google"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/harvester"
	linodeDefaults "github.com/rancher/tfp-automation/framework/set/defaults/providers/linode"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/vsphere"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	rke2ServerOne   = "server1"
	rke2ServerTwo   = "server2"
	rke2ServerThree = "server3"
	token           = "token"
)

// CreateRKE2Cluster is a helper function that will create the RKE2 cluster.
func CreateRKE2Cluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.RKE2KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	serverScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/rke2/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/rke2/add-servers.sh")

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
	nullResourceBlock := rootBody.AppendNewBlock(general.Resource, []string{general.NullResource, host})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(general.Provisioner, []string{general.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(general.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	connectionBlockBody.SetAttributeValue(general.Host, cty.StringVal(instance))
	connectionBlockBody.SetAttributeValue(general.Type, cty.StringVal(general.Ssh))
	switch terraformConfig.Provider {
	case aws.Aws:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

		keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	case azure.Azure, providers.AKS:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.AzureConfig.SSHUser))

		keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	case google.Google:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.GoogleConfig.SSHUser))

		keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	case linodeDefaults.Linode:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(linode.RootUser))
		connectionBlockBody.SetAttributeValue(general.Password, cty.StringVal(terraformConfig.LinodeConfig.LinodeRootPass))
	case harvester.Harvester:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.HarvesterConfig.SSHUser))

		keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	case vsphere.Vsphere:
		connectionBlockBody.SetAttributeValue(general.User, cty.StringVal(terraformConfig.VsphereConfig.VsphereUser))

		keyPathExpression := general.File + `("` + terraformConfig.PrivateKeyPath + `")`
		keyPath := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
		}

		connectionBlockBody.SetAttributeRaw(general.PrivateKey, keyPath)
	}

	rootBody.AppendNewline()

	return nullResourceBlockBody, provisionerBlockBody
}

// createRKE2Server is a helper function that will create the RKE2 server.
func createRKE2Server(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP,
	rke2Token string, script []byte) {
	_, provisionerBlockBody := SSHNullResource(rootBody, terraformConfig, rke2ServerOnePublicIP, rke2ServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " + rke2Token + " " + terraformConfig.CNI + " " +
		terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + "'"

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
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

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.RKE2Version + " " +
			rke2ServerOnePrivateIP + " " + instance + " " + rke2Token + " " + terraformConfig.CNI + " " + terraformConfig.Standalone.RegistryUsername + " " +
			terraformConfig.Standalone.RegistryPassword + "'"

		provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("printf '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
			cty.StringVal(command),
		}))

		dependsOnServer := `[` + general.NullResource + `.` + rke2ServerOne + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
	}
}
