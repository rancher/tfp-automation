package imported

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// CreateDualStackRKE2K3SImportedCluster is a helper function that will create the RKE2/K3S cluster to be imported into Rancher.
func CreateDualStackRKE2K3SImportedCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	linuxNodeNames, serverNodeNames, agentNodeNames []string, nodePublicIPs, nodePrivateIPs map[string]string, token string) error {
	if len(serverNodeNames) == 0 {
		return nil
	}

	var serverScriptPath, addServersScriptPath, addAgentsScriptPath string

	userDir, _ := rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
		serverScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/k3s/init-server.sh")
		addServersScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/k3s/add-servers.sh")
		addAgentsScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/k3s/add-agents.sh")
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
		serverScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/rke2/init-server.sh")
		addServersScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/rke2/add-servers.sh")
		addAgentsScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/dualstack/rke2/add-agents.sh")
	}

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return err
	}

	newServersScriptContent, err := os.ReadFile(addServersScriptPath)
	if err != nil {
		return err
	}

	agentsScriptContent, err := os.ReadFile(addAgentsScriptPath)
	if err != nil {
		return err
	}

	bootstrapNodeName := serverNodeNames[0]
	serverOnePublicIP := nodePublicIPs[bootstrapNodeName]
	serverOnePrivateIP := nodePrivateIPs[bootstrapNodeName]

	createImportedDualStackRKE2K3SServer(rootBody, terraformConfig, linuxNodeNames, bootstrapNodeName, serverOnePublicIP, serverOnePrivateIP, token, serverOneScriptContent)
	addImportedDualStackRKE2K3SServerNodes(rootBody, terraformConfig, linuxNodeNames, serverNodeNames[1:], serverOnePrivateIP, nodePublicIPs, token, newServersScriptContent)
	addImportedDualStackRKE2K3SAgentNodes(rootBody, terraformConfig, linuxNodeNames, agentNodeNames, serverOnePrivateIP, nodePublicIPs, token, agentsScriptContent)

	return nil
}

// createImportedDualStackRKE2K3SServer is a helper function that will create the server to be imported into Rancher.
func createImportedDualStackRKE2K3SServer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, linuxNodeNames []string,
	bootstrapNodeName, serverOnePublicIP, serverOnePrivateIP, token string, script []byte) {
	copyScriptName := terraformConfig.ResourcePrefix + copyScript + bootstrapNodeName
	_, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, copyScriptName, linuxNodeNames)

	var version, command string

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
		version = terraformConfig.Standalone.K3SVersion

		command = "/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			version + " " + serverOnePrivateIP + " " + token + " " + terraformConfig.Standalone.RegistryUsername + " " +
			terraformConfig.Standalone.RegistryPassword + " " + terraformConfig.AWSConfig.ClusterCIDR + " " +
			terraformConfig.AWSConfig.ServiceCIDR
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
		version = terraformConfig.Standalone.RKE2Version

		command = "/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			version + " " + serverOnePrivateIP + " " + token + " " + terraformConfig.CNI + " " +
			terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
			terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
	}

	// For imported clusters, need to first put the script on the machine before running it.
	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(script) + "' > /tmp/init-server.sh"),
		cty.StringVal("chmod +x /tmp/init-server.sh"),
	}))

	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster
	nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, createClusterName, linuxNodeNames)

	provisionerBlockBody.SetAttributeRaw(general.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	dependsOnServer := `[` + general.NullResource + `.` + copyScriptName + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
}

// addImportedDualStackRKE2K3SServerNodes is a helper function that will add additional server nodes to the initial server.
func addImportedDualStackRKE2K3SServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, linuxNodeNames, serverNodeNames []string,
	serverOnePrivateIP string, nodePublicIPs map[string]string, token string, script []byte) {
	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster

	for _, nodeName := range serverNodeNames {
		resourceName := nodeName
		instance := nodePublicIPs[nodeName]
		nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, resourceName, linuxNodeNames)

		var version, command string

		if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
			version = terraformConfig.Standalone.K3SVersion

			command = "/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				version + " " + serverOnePrivateIP + " " + instance + " " + token + " " +
				terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " +
				terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
		} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
			version = terraformConfig.Standalone.RKE2Version

			command = "/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + version + " " +
				serverOnePrivateIP + " " + instance + " " + token + " " + terraformConfig.CNI + " " + terraformConfig.Standalone.RegistryUsername + " " +
				terraformConfig.Standalone.RegistryPassword + " " + terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
		}

		provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("echo '" + string(script) + "' > /tmp/add-servers.sh"),
			cty.StringVal("chmod +x /tmp/add-servers.sh"),
		}))

		dependsOnServer := `[` + general.NullResource + `.` + createClusterName + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

		nullResourceBlockBody, provisionerBlockBody = nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, addServer+"_"+resourceName, linuxNodeNames)

		provisionerBlockBody.SetAttributeRaw(general.Inline, hclwrite.Tokens{
			{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
			{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
		})

		importClusterName := resourceName
		dependsOnServer = `[` + general.NullResource + `.` + importClusterName + `]`

		server = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
	}

}

func addImportedDualStackRKE2K3SAgentNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, linuxNodeNames, agentNodeNames []string,
	serverOnePrivateIP string, nodePublicIPs map[string]string, token string, script []byte) {
	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster

	for _, nodeName := range agentNodeNames {
		resourceName := nodeName
		instance := nodePublicIPs[nodeName]
		nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, resourceName, linuxNodeNames)

		var version, command string

		if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
			version = terraformConfig.Standalone.K3SVersion

			command = "/tmp/add-agents.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				version + " " + serverOnePrivateIP + " " + instance + " " + token + " " +
				terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword
		} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
			version = terraformConfig.Standalone.RKE2Version

			command = "/tmp/add-agents.sh " + terraformConfig.Standalone.OSUser + " " + version + " " +
				serverOnePrivateIP + " " + instance + " " + token + " " + terraformConfig.CNI + " " + terraformConfig.Standalone.RegistryUsername + " " +
				terraformConfig.Standalone.RegistryPassword + " " + terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
		}

		provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
			cty.StringVal("echo '" + string(script) + "' > /tmp/add-agents.sh"),
			cty.StringVal("chmod +x /tmp/add-agents.sh"),
		}))

		dependsOnServer := `[` + general.NullResource + `.` + createClusterName + `]`
		server := hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

		nullResourceBlockBody, provisionerBlockBody = nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, addAgent+"_"+resourceName, linuxNodeNames)

		provisionerBlockBody.SetAttributeRaw(general.Inline, hclwrite.Tokens{
			{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
			{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
			{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
		})

		importClusterName := resourceName
		dependsOnServer = `[` + general.NullResource + `.` + importClusterName + `]`

		server = hclwrite.Tokens{
			{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
		}

		nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
	}

}
