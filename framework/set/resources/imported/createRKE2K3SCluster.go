package imported

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	addServer     = "add_server"
	createCluster = "create_cluster"
	serverOne     = "server1"
	serverTwo     = "server2"
	serverThree   = "server3"
	token         = "token"
)

// CreateRKE2K3SImportedCluster is a helper function that will create the RKE2/K3S cluster to be imported into Rancher.
func CreateRKE2K3SImportedCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, serverOnePublicDNS, serverOnePrivateIP,
	serverTwoPublicDNS, serverThreePublicDNS string) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	var serverScriptPath, newServersScriptPath string

	if terraformConfig.Module == modules.ImportEC2K3s {
		serverScriptPath = filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/k3s/init-server.sh")
		newServersScriptPath = filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/k3s/add-servers.sh")
	} else {
		serverScriptPath = filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/rke2/init-server.sh")
		newServersScriptPath = filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/rke2/add-servers.sh")
	}

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return
	}

	newServersScriptContent, err := os.ReadFile(newServersScriptPath)
	if err != nil {
		return
	}

	token := namegen.AppendRandomString(token)

	createImportedRKE2K3SServer(rootBody, terraformConfig, serverOnePublicDNS, serverOnePrivateIP, token, serverOneScriptContent)
	addImportedRKE2K3SServerNodes(rootBody, terraformConfig, serverOnePrivateIP, serverTwoPublicDNS, serverThreePublicDNS, token, newServersScriptContent)
}

// CreateImportedNullResource is a helper function that will create the null_resource for the cluster.
func CreateImportedNullResource(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, instance, host string) (*hclwrite.Body, *hclwrite.Body) {
	nullResourceBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.NullResource, host})
	nullResourceBlockBody := nullResourceBlock.Body()

	provisionerBlock := nullResourceBlockBody.AppendNewBlock(defaults.Provisioner, []string{defaults.RemoteExec})
	provisionerBlockBody := provisionerBlock.Body()

	connectionBlock := provisionerBlockBody.AppendNewBlock(defaults.Connection, nil)
	connectionBlockBody := connectionBlock.Body()

	hostExpression := `"` + instance + `"`

	newHost := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(hostExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.Host, newHost)

	connectionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(defaults.Ssh))
	connectionBlockBody.SetAttributeValue(defaults.User, cty.StringVal(terraformConfig.AWSConfig.AWSUser))

	keyPathExpression := defaults.File + `("` + terraformConfig.PrivateKeyPath + `")`
	keyPath := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(keyPathExpression)},
	}

	connectionBlockBody.SetAttributeRaw(defaults.PrivateKey, keyPath)

	dependsOnServer := `[` + defaults.AwsInstance + `.` + serverOne + `, ` + defaults.AwsInstance + `.` + serverTwo + `, ` + defaults.AwsInstance + `.` + serverThree + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	rootBody.AppendNewline()

	return nullResourceBlockBody, provisionerBlockBody
}

// createImportedRKE2K3SServer is a helper function that will create the server to be imported into Rancher.
func createImportedRKE2K3SServer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, serverOnePublicDNS, serverOnePrivateIP,
	token string, script []byte) {
	_, provisionerBlockBody := CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicDNS, serverOne)

	var version string

	if terraformConfig.Module == modules.ImportEC2K3s {
		version = terraformConfig.Standalone.K3SVersion
	} else {
		version = terraformConfig.Standalone.RKE2Version
	}

	command := "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
		version + " " + serverOnePrivateIP + " " + token + "'"

	// For imported clusters, need to first put the script on the machine before running it.
	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
	}))

	nullResourceBlockBody, provisionerBlockBody := CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicDNS, createCluster)

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	dependsOnServer := `[` + defaults.NullResource + `.` + serverOne + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
}

// addImportedServerNodes is a helper function that will add additional server nodes to the initial server.
func addImportedRKE2K3SServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, serverOnePrivateIP, serverTwoPublicDNS,
	serverThreePublicDNS, token string, script []byte) {
	instances := []string{serverTwoPublicDNS, serverThreePublicDNS}
	hosts := []string{serverTwo, serverThree}

	for i, instance := range instances {
		host := hosts[i]
		nullResourceBlockBody, provisionerBlockBody := CreateImportedNullResource(rootBody, terraformConfig, instance, host)

		var version string

		if terraformConfig.Module == modules.ImportEC2K3s {
			version = terraformConfig.Standalone.K3SVersion
		} else {
			version = terraformConfig.Standalone.RKE2Version
		}

		command := "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			version + " " + serverOnePrivateIP + " " + token + "'"

		provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("echo '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
		}))

		dependsOnServer := `[` + defaults.NullResource + `.` + serverOne + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

		nullResourceBlockBody, provisionerBlockBody = CreateImportedNullResource(rootBody, terraformConfig, instance, addServer+"_"+host)

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
			{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
			{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
		})

		dependsOnServer = `[` + defaults.NullResource + `.` + createCluster + `]`

		server = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
	}
}
