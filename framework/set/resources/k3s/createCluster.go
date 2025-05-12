package k3s

import (
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
	k3sServerOne   = "k3s_server1"
	k3sServerTwo   = "k3s_server2"
	k3sServerThree = "k3s_server3"
	token          = "token"
)

// CreateK3SCluster is a helper function that will create the K3S cluster.
func CreateK3SCluster(file *os.File, newFile *hclwrite.File, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	k3sServerOnePublicDNS, k3sServerOnePrivateIP, k3sServerTwoPublicDNS, k3sServerThreePublicDNS string) (*os.File, error) {
	userDir, _ := rancher2.SetKeyPath(keypath.K3sKeyPath, terraformConfig.Provider)

	serverScriptPath := filepath.Join(userDir, "src/github.com/rancher/tfp-automation/framework/set/resources/k3s/init-server.sh")
	newServersScriptPath := filepath.Join(userDir, "src/github.com/rancher/tfp-automation/framework/set/resources/k3s/add-servers.sh")

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return nil, err
	}

	newServersScriptContent, err := os.ReadFile(newServersScriptPath)
	if err != nil {
		return nil, err
	}

	k3sToken := namegen.AppendRandomString(token)

	CreateK3SServer(rootBody, terraformConfig, k3sServerOnePublicDNS, k3sServerOnePrivateIP, k3sToken, serverOneScriptContent)
	AddK3SServerNodes(rootBody, terraformConfig, k3sServerOnePrivateIP, k3sServerTwoPublicDNS, k3sServerThreePublicDNS, k3sToken, newServersScriptContent)

	_, err = file.Write(newFile.Bytes())
	if err != nil {
		logrus.Infof("Failed to append configurations to main.tf file. Error: %v", err)
		return nil, err
	}

	return file, nil
}

// CreateK3SServer is a helper function that will create the K3S server.
func CreateK3SServer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, k3sServerOnePublicDNS, k3sServerOnePrivateIP,
	k3sToken string, script []byte) {
	_, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, k3sServerOnePublicDNS, k3sServerOne)

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		terraformConfig.Standalone.K3SVersion + " " + k3sServerOnePrivateIP + " " + k3sToken + "'"

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("printf '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
		cty.StringVal(command),
	}))
}

// AddK3SServerNodes is a helper function that will add additional K3s server nodes to the initial K3s server.
func AddK3SServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, k3sServerOnePrivateIP, k3sServerTwoPublicDNS,
	k3sServerThreePublicDNS, k3sToken string, script []byte) {
	instances := []string{k3sServerTwoPublicDNS, k3sServerThreePublicDNS}
	hosts := []string{k3sServerTwo, k3sServerThree}

	for i, instance := range instances {
		host := hosts[i]
		nullResourceBlockBody, provisionerBlockBody := rke2.SSHNullResource(rootBody, terraformConfig, instance, host)

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			terraformConfig.Standalone.K3SVersion + " " + k3sServerOnePrivateIP + " " + k3sToken + "'"

		provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("printf '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
			cty.StringVal(command),
		}))

		dependsOnServer := `[` + defaults.NullResource + `.` + k3sServerOne + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
	}
}
