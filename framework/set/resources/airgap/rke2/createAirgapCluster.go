package rke2

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/rke2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const (
	rke2Bastion     = "rke2_bastion"
	rke2ServerOne   = "rke2_server1"
	rke2ServerTwo   = "rke2_server2"
	rke2ServerThree = "rke2_server3"
	token           = "token"
)

// CreateAirgapRKE2Cluster is a helper function that will create the RKE2 cluster.
func CreateAirgapRKE2Cluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	rke2BastionPublicDNS, registryPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP string) (*os.File, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	bastionScriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/airgap/rke2/bastion.sh")
	serverScriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/airgap/rke2/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/airgap/rke2/add-servers.sh")

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

	_, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2BastionPublicDNS, rke2Bastion)

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(bastionScriptContent) + "' > /tmp/bastion.sh"),
		cty.StringVal("chmod +x /tmp/bastion.sh"),
		cty.StringVal("bash -c '/tmp/bastion.sh " + terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " +
			rke2ServerTwoPrivateIP + " " + rke2ServerThreePrivateIP + " " + terraformConfig.Standalone.OSUser + " " + encodedPEMFile + "'"),
	}))

	rke2Token := namegen.AppendRandomString(token)

	createAirgappedRKE2Server(rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2Token, registryPublicDNS, serverOneScriptContent)
	addAirgappedRKE2ServerNodes(rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPrivateIP, rke2ServerThreePrivateIP, rke2Token, registryPublicDNS, newServersScriptContent)

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// createAirgappedRKE2Server is a helper function that will create the RKE2 server.
func createAirgappedRKE2Server(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP,
	rke2Token, registryPublicDNS string, script []byte) {
	nullResourceBlockBody, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2BastionPublicDNS, rke2ServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		rke2ServerOnePrivateIP + " " + rke2Token + " " + registryPublicDNS + " " + terraformConfig.Standalone.RancherImage + " " +
		terraformConfig.Standalone.RancherTagVersion

	if terraformConfig.Standalone.RancherAgentImage != "" {
		command += " " + terraformConfig.Standalone.RancherAgentImage
	}

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

// addAirgappedRKE2ServerNodes is a helper function that will add additional RKE2 server nodes to the initial RKE2 airgapped server.
func addAirgappedRKE2ServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2BastionPublicDNS, rke2ServerOnePrivateIP, rke2ServerTwoPublicDNS,
	rke2ServerThreePublicDNS, rke2Token, registryPublicDNS string, script []byte) {
	instances := []string{rke2ServerTwoPublicDNS, rke2ServerThreePublicDNS}
	hosts := []string{rke2ServerTwo, rke2ServerThree}

	for i, instance := range instances {
		host := hosts[i]
		nullResourceBlockBody, provisionerBlockBody := rke2.CreateNullResource(rootBody, terraformConfig, rke2BastionPublicDNS, host)

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			rke2ServerOnePrivateIP + " " + instance + " " + rke2Token + " " + registryPublicDNS + " " +
			terraformConfig.Standalone.RancherImage + " " + terraformConfig.Standalone.RancherTagVersion

		if terraformConfig.Standalone.RancherAgentImage != "" {
			command += " " + terraformConfig.Standalone.RancherAgentImage
		}

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
