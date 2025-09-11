package rke2

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	rke2Bastion     = "bastion"
	rke2ServerOne   = "server1"
	rke2ServerTwo   = "server2"
	rke2ServerThree = "server3"
	token           = "token"
)

// CreateIPv6RKE2Cluster is a helper function that will create the RKE2 cluster.
func CreateIPv6RKE2Cluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2BastionPublicIP, rke2ServerOnePublicIP, rke2ServerTwoPublicIP,
	rke2ServerThreePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.IPv6KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	bastionScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/airgap/rke2/bastion.sh")
	serverScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/rke2/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/rke2/add-servers.sh")

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

	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2BastionPublicIP, rke2Bastion)

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(bastionScriptContent) + "' > /tmp/bastion.sh"),
		cty.StringVal("chmod +x /tmp/bastion.sh"),
		cty.StringVal("bash -c '/tmp/bastion.sh " + terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " +
			rke2ServerTwoPrivateIP + " " + rke2ServerThreePrivateIP + " " + terraformConfig.Standalone.OSUser + " " + encodedPEMFile + "'"),
	}))

	rke2Token := namegen.AppendRandomString(token)

	createIPv6RKE2Server(rootBody, terraformConfig, rke2BastionPublicIP, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2Token, serverOneScriptContent)
	addIPv6RKE2ServerNodes(rootBody, terraformConfig, rke2BastionPublicIP, rke2ServerOnePublicIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP,
		rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP, rke2Token, newServersScriptContent)

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// createIPv6RKE2Server is a helper function that will create the RKE2 server.
func createIPv6RKE2Server(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2BastionPublicIP, rke2ServerOnePublicIP,
	rke2ServerOnePrivateIP, rke2Token string, script []byte) {
	nullResourceBlockBody, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2BastionPublicIP, rke2ServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		rke2ServerOnePublicIP + " " + rke2ServerOnePrivateIP + " " + terraformConfig.Standalone.RancherHostname + " " +
		terraformConfig.CNI + " " + terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
		rke2Token + " " + terraformConfig.Standalone.RancherImage + " " + terraformConfig.Standalone.RancherTagVersion + " " +
		terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR

	command += " || true'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
		cty.StringVal(command),
	}))

	dependsOnServer := `[` + defaults.NullResource + `.` + rke2Bastion + `]`
	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
}

// addIPv6RKE2ServerNodes is a helper function that will add additional RKE2 server nodes to the initial RKE2 IPv6 server.
func addIPv6RKE2ServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2BastionPublicIP, rke2ServerOnePublicIP,
	rke2ServerTwoPublicIP, rke2ServerThreePublicIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP, rke2Token string, script []byte) {
	privateIPInstances := []string{rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP}
	publicIPInstances := []string{rke2ServerTwoPublicIP, rke2ServerThreePublicIP}
	hosts := []string{rke2ServerTwo, rke2ServerThree}

	for i, privateInstance := range privateIPInstances {
		publicIPInstance := publicIPInstances[i]
		host := hosts[i]

		nullResourceBlockBody, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2BastionPublicIP, host)

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			rke2ServerOnePublicIP + " " + publicIPInstance + " " + privateInstance + " " + terraformConfig.Standalone.RancherHostname + " " +
			terraformConfig.CNI + " " + terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
			rke2Token + " " + terraformConfig.Standalone.RancherImage + " " + terraformConfig.Standalone.RancherTagVersion + " " +
			terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR

		command += " || true'"

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
