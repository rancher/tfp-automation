package rke2

import (
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
	rke2ServerOne   = "server1"
	rke2ServerTwo   = "server2"
	rke2ServerThree = "server3"
	token           = "token"
)

// CreateRKE2Cluster is a helper function that will create the RKE2 cluster.
func CreateRKE2Cluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP, rke2ServerTwoPublicIP, rke2ServerThreePublicIP string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.RKE2KeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	serverScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/rke2/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/rke2/add-servers.sh")

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

// createRKE2Server is a helper function that will create the RKE2 server.
func createRKE2Server(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, rke2ServerOnePublicIP, rke2ServerOnePrivateIP,
	rke2Token string, script []byte) {
	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, rke2ServerOnePublicIP, rke2ServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		terraformConfig.Standalone.RKE2Version + " " + rke2ServerOnePrivateIP + " " + rke2Token + " " + terraformConfig.CNI + " " +
		terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
		terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR + "'"

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
		nullResourceBlockBody, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, instance, host)

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.RKE2Version + " " +
			rke2ServerOnePrivateIP + " " + instance + " " + rke2Token + " " + terraformConfig.CNI + " " + terraformConfig.Standalone.RegistryUsername + " " +
			terraformConfig.Standalone.RegistryPassword + " " + terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR + "'"

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
