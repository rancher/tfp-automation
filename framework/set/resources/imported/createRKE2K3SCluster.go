package imported

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported/nullresource"
	"github.com/zclconf/go-cty/cty"
)

const (
	addServer      = "add_server"
	addWindowsNode = "add_windows_node"
	copyScript     = "_copy_script_"
	createCluster  = "create_cluster"
	serverOne      = "server1"
	serverTwo      = "server2"
	serverThree    = "server3"
	windowsServer  = "windows_server"
	token          = "token"
)

// CreateRKE2K3SImportedCluster is a helper function that will create the RKE2/K3S cluster to be imported into Rancher.
func CreateRKE2K3SImportedCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	serverOnePublicIP, serverOnePrivateIP, serverTwoPublicIP, serverThreePublicIP, token string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var serverScriptPath, newServersScriptPath string

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, defaults.Import) {
		serverScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/k3s/init-server.sh")
		newServersScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/k3s/add-servers.sh")
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, defaults.Import) {
		serverScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/rke2/init-server.sh")
		newServersScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/rke2/add-servers.sh")
	}

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return err
	}

	newServersScriptContent, err := os.ReadFile(newServersScriptPath)
	if err != nil {
		return err
	}

	createImportedRKE2K3SServer(rootBody, terraformConfig, serverOnePublicIP, serverOnePrivateIP, token, serverOneScriptContent)
	addImportedRKE2K3SServerNodes(rootBody, terraformConfig, serverOnePrivateIP, serverTwoPublicIP, serverThreePublicIP, token, newServersScriptContent)

	return nil
}

// createImportedRKE2K3SServer is a helper function that will create the server to be imported into Rancher.
func createImportedRKE2K3SServer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, serverOnePublicIP, serverOnePrivateIP,
	token string, script []byte) {
	copyScriptName := terraformConfig.ResourcePrefix + copyScript + serverOne
	_, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, copyScriptName)

	var version, command string

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, defaults.Import) {
		version = terraformConfig.Standalone.K3SVersion

		command = "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			terraformConfig.Standalone.K3SVersion + " " + serverOnePrivateIP + " " + token + "'"
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, defaults.Import) {
		version = terraformConfig.Standalone.RKE2Version

		command = "bash -c '/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			version + " " + serverOnePrivateIP + " " + token + " " + terraformConfig.CNI + " || true'"
	}

	// For imported clusters, need to first put the script on the machine before running it.
	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
	}))

	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster
	nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, createClusterName)

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	dependsOnServer := `[` + defaults.NullResource + `.` + copyScriptName + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
}

// addImportedServerNodes is a helper function that will add additional server nodes to the initial server.
func addImportedRKE2K3SServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, serverOnePrivateIP, serverTwoPublicIP,
	serverThreePublicIP, token string, script []byte) {
	instances := []string{serverTwoPublicIP, serverThreePublicIP}
	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster
	resourceNames := []string{serverTwo, serverThree}

	for i, instance := range instances {
		resourceName := terraformConfig.ResourcePrefix + `_` + resourceNames[i]
		nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, resourceName)

		var version, command string

		if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, defaults.Import) {
			version = terraformConfig.Standalone.K3SVersion

			command = "bash -c '/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				terraformConfig.Standalone.K3SVersion + " " + serverOnePrivateIP + " " + instance + " " + token + "'"
		} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, defaults.Import) {
			version = terraformConfig.Standalone.RKE2Version

			command = "bash -c '/tmp/add-servers.sh " + version + " " + serverOnePrivateIP + " " + instance + " " + token + " " +
				terraformConfig.CNI + " || true'"
		}

		provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("echo '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
		}))

		dependsOnServer := `[` + defaults.NullResource + `.` + createClusterName + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

		nullResourceBlockBody, provisionerBlockBody = nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, addServer+"_"+resourceName)

		provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
			{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
			{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
		})

		importClusterName := terraformConfig.ResourcePrefix + `_` + resourceNames[i]
		dependsOnServer = `[` + defaults.NullResource + `.` + importClusterName + `]`

		server = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
	}
}
