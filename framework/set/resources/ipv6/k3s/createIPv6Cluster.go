package k3s

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	k3sBastion     = "bastion"
	k3sServerOne   = "server1"
	k3sServerTwo   = "server2"
	k3sServerThree = "server3"
	token          = "token"
)

// CreateIPv6K3SCluster is a helper function that will create the K3S cluster.
func CreateIPv6K3SCluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, k3sBastionPublicIP, k3sServerOnePublicIP, k3sServerTwoPublicIP,
	k3sServerThreePublicIP, k3sServerOnePrivateIP, k3sServerTwoPrivateIP, k3sServerThreePrivateIP string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.IPv6KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	bastionScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/airgap/k3s/bastion.sh")
	serverScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/k3s/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/k3s/add-servers.sh")

	bastionScriptContent, err := os.ReadFile(bastionScriptPath)
	if err != nil {
		return nil, err
	}

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return nil, err
	}

	newServersScriptContent, err := os.ReadFile(newServersScriptPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	encodedPEMFile := base64.StdEncoding.EncodeToString([]byte(privateKey))

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, k3sBastionPublicIP, k3sBastion)

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(bastionScriptContent) + "' > /tmp/bastion.sh"),
		cty.StringVal("chmod +x /tmp/bastion.sh"),
		cty.StringVal("bash -c '/tmp/bastion.sh " + terraformConfig.Standalone.K3SVersion + " " + k3sServerOnePrivateIP + " " +
			k3sServerTwoPrivateIP + " " + k3sServerThreePrivateIP + " " + terraformConfig.Standalone.OSUser + " " + encodedPEMFile + "'"),
	}))

	k3sToken := namegen.AppendRandomString(token)

	createIPv6K3SServer(rootBody, terraformConfig, k3sBastionPublicIP, k3sServerOnePublicIP, k3sServerOnePrivateIP, k3sToken, serverOneScriptContent)
	addIPv6K3SServerNodes(rootBody, terraformConfig, k3sBastionPublicIP, k3sServerOnePublicIP, k3sServerTwoPrivateIP, k3sServerThreePrivateIP, k3sToken,
		newServersScriptContent)

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// createIPv6K3SServer is a helper function that will create the K3S server.
func createIPv6K3SServer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, k3sBastionPublicIP, k3sServerOnePublicIP,
	k3sServerOnePrivateIP, k3sToken string, script []byte) {
	nullResourceBlockBody, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, k3sBastionPublicIP, k3sServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		k3sServerOnePublicIP + " " + k3sServerOnePrivateIP + " " + terraformConfig.Standalone.RancherHostname + " " +
		terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
		k3sToken + " " + terraformConfig.Standalone.RancherImage + " " + terraformConfig.Standalone.RancherTagVersion + " " +
		terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR + "'"

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
		cty.StringVal(command),
	}))

	dependsOnServer := `[` + general.NullResource + `.` + k3sBastion + `]`
	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
}

// addIPv6K3SServerNodes is a helper function that will add additional K3S server nodes to the initial K3S IPv6 server.
func addIPv6K3SServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, k3sBastionPublicIP, k3sServerOnePublicIP,
	k3sServerTwoPrivateIP, k3sServerThreePrivateIP, k3sToken string, script []byte) {
	privateIPInstances := []string{k3sServerTwoPrivateIP, k3sServerThreePrivateIP}
	hosts := []string{k3sServerTwo, k3sServerThree}

	for i, privateInstance := range privateIPInstances {
		host := hosts[i]

		nullResourceBlockBody, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, k3sBastionPublicIP, host)

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			k3sServerOnePublicIP + " " + privateInstance + " " + terraformConfig.Standalone.RancherHostname + " " +
			terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
			k3sToken + " " + terraformConfig.Standalone.RancherImage + " " + terraformConfig.Standalone.RancherTagVersion + " " +
			terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR + "'"

		provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("printf '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
			cty.StringVal(command),
		}))

		dependsOnServer := `[` + general.NullResource + `.` + k3sServerOne + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
	}
}
