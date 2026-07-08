package imported

import (
	"encoding/base64"
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

// CreateIPv6RKE2K3SImportedCluster is a helper function that will create the IPv6 RKE2/K3S cluster to be imported into Rancher.
func CreateIPv6RKE2K3SImportedCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	linuxNodeNames, serverNodeNames, agentNodeNames []string, nodePublicIPs, nodePublicIPv6s, nodePrivateIPs map[string]string,
	token string) error {
	if len(serverNodeNames) == 0 {
		return nil
	}

	var bastionScriptPath, serverScriptPath, addServersScriptPath, addAgentsScriptPath string

	userDir, _ := rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
		bastionScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/k3s/imported-bastion.sh")
		serverScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/k3s/init-server.sh")
		addServersScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/k3s/add-servers.sh")
		addAgentsScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/k3s/add-agents.sh")
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
		bastionScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/rke2/imported-bastion.sh")
		serverScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/rke2/init-server.sh")
		addServersScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/rke2/add-servers.sh")
		addAgentsScriptPath = filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/resources/ipv6/rke2/add-agents.sh")
	}

	bastionScriptContent, err := os.ReadFile(bastionScriptPath)
	if err != nil {
		return err
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

	privateKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	if err != nil {
		return err
	}

	encodedPEMFile := base64.StdEncoding.EncodeToString([]byte(privateKey))

	bootstrapNodeName := serverNodeNames[0]
	serverOnePublicIP := nodePublicIPs[bootstrapNodeName]
	serverOnePublicIPv6 := nodePublicIPv6s[bootstrapNodeName]
	serverOnePrivateIP := nodePrivateIPs[bootstrapNodeName]

	// Needed to ensure that both additional server and agent nodes are accounted for in the bastion node setup.
	additionalServerPrivateIPs := make([]string, 0, len(serverNodeNames)-1+len(agentNodeNames))
	for _, serverNodeName := range serverNodeNames[1:] {
		additionalServerPrivateIPs = append(additionalServerPrivateIPs, nodePrivateIPs[serverNodeName])
	}

	for _, agentNodeName := range agentNodeNames {
		additionalServerPrivateIPs = append(additionalServerPrivateIPs, nodePrivateIPs[agentNodeName])
	}

	bastionNodeSetup(rootBody, terraformConfig, terratestConfig, linuxNodeNames, bastionScriptContent, encodedPEMFile, serverOnePublicIP, serverOnePrivateIP, additionalServerPrivateIPs)
	createImportedIPv6RKE2K3SServer(rootBody, terraformConfig, linuxNodeNames, bootstrapNodeName, serverOnePublicIP, serverOnePublicIPv6, serverOnePrivateIP, token, serverOneScriptContent)
	addImportedIPv6RKE2K3SServerNodes(rootBody, terraformConfig, linuxNodeNames, serverNodeNames[1:], serverOnePublicIP, serverOnePublicIPv6, serverOnePrivateIP, nodePublicIPs, nodePublicIPv6s, nodePrivateIPs, token, newServersScriptContent)
	addImportedIPv6RKE2K3SAgentNodes(rootBody, terraformConfig, linuxNodeNames, agentNodeNames, serverOnePublicIPv6, serverOnePrivateIP, nodePublicIPs, nodePublicIPv6s, nodePrivateIPs, token, agentsScriptContent)

	return nil
}

// bastionNodeSetup is a helper function that will set up the bastion node for the imported cluster.
func bastionNodeSetup(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	linuxNodeNames []string, script []byte, encodedPEMFile, serverOnePublicIP, serverOnePrivateIP string, additionalServerPrivateIPs []string) {
	bastionNodeName := terraformConfig.ResourcePrefix + `-` + `bastion`
	copyScriptName := terraformConfig.ResourcePrefix + copyScript + bastionNodeName
	_, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, copyScriptName, linuxNodeNames)

	var version, additionalServers string

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
		version = terraformConfig.Standalone.K3SVersion
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
		version = terraformConfig.Standalone.RKE2Version
	}

	for _, ip := range additionalServerPrivateIPs {
		additionalServers += " " + ip
	}

	command := "/tmp/imported-bastion.sh " + version + " " + serverOnePrivateIP + additionalServers + " " +
		terraformConfig.Standalone.OSUser + " " + encodedPEMFile

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(script) + "' > /tmp/imported-bastion.sh"),
		cty.StringVal("chmod +x /tmp/imported-bastion.sh"),
	}))

	nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, bastionNodeName, linuxNodeNames)

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

// createImportedRKE2K3SServer is a helper function that will create the server to be imported into Rancher.
func createImportedIPv6RKE2K3SServer(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, linuxNodeNames []string,
	bootstrapNodeName, serverOnePublicIP, serverOnePublicIPv6, serverOnePrivateIP, token string, script []byte) {
	bastionNodeName := terraformConfig.ResourcePrefix + `-` + `bastion`
	copyScriptName := terraformConfig.ResourcePrefix + copyScript + bootstrapNodeName
	_, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, serverOnePublicIP, copyScriptName, linuxNodeNames)

	var command string

	if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
		command = "/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			terraformConfig.Standalone.K3SVersion + " " + serverOnePublicIPv6 + " " + serverOnePrivateIP + " " +
			terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + token + " " +
			terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
	} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
		command = "/tmp/init-server.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			serverOnePublicIPv6 + " " + serverOnePrivateIP + " " + terraformConfig.CNI + " " +
			terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + token + " " +
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

	dependsOnServer := `[` + general.NullResource + `.` + bastionNodeName + `]`
	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)
}

func addImportedIPv6RKE2K3SServerNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, linuxNodeNames, serverNodeNames []string,
	serverOnePublicIP, serverOnePublicIPv6, serverOnePrivateIP string, nodePublicIPs, nodePublicIPv6s, nodePrivateIPs map[string]string, token string, script []byte) {
	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster

	for _, nodeName := range serverNodeNames {
		resourceName := nodeName
		instance := nodePublicIPs[nodeName]
		publicInstance := nodePublicIPv6s[nodeName]
		privateInstance := nodePrivateIPs[nodeName]

		nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, resourceName, linuxNodeNames)

		var command string

		if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
			command = "/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				terraformConfig.Standalone.K3SVersion + " " + serverOnePublicIPv6 + " " + serverOnePrivateIP + " " + privateInstance + " " +
				terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + token + " " +
				terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
		} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
			command = "/tmp/add-servers.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				serverOnePublicIPv6 + " " + serverOnePrivateIP + " " + publicInstance + " " + privateInstance + " " + terraformConfig.CNI + " " +
				terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + token + " " +
				terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
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

func addImportedIPv6RKE2K3SAgentNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, linuxNodeNames, agentNodeNames []string,
	serverOnePublicIPv6, serverOnePrivateIP string, nodePublicIPs, nodePublicIPv6s, nodePrivateIPs map[string]string, token string, script []byte) {
	createClusterName := terraformConfig.ResourcePrefix + `_` + createCluster

	for _, nodeName := range agentNodeNames {
		resourceName := nodeName
		instance := nodePublicIPs[nodeName]
		publicInstance := nodePublicIPv6s[nodeName]
		privateInstance := nodePrivateIPs[nodeName]

		nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, instance, resourceName, linuxNodeNames)

		var command string

		if strings.Contains(terraformConfig.Module, clustertypes.K3S) && strings.Contains(terraformConfig.Module, general.Import) {
			command = "/tmp/add-agents.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				terraformConfig.Standalone.K3SVersion + " " + serverOnePublicIPv6 + " " + serverOnePrivateIP + " " + privateInstance + " " + token + " " +
				terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword
		} else if strings.Contains(terraformConfig.Module, clustertypes.RKE2) && strings.Contains(terraformConfig.Module, general.Import) {
			command = "/tmp/add-agents.sh " + terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
				serverOnePublicIPv6 + " " + serverOnePrivateIP + " " + publicInstance + " " + privateInstance + " " + terraformConfig.CNI + " " +
				terraformConfig.Standalone.RegistryUsername + " " + terraformConfig.Standalone.RegistryPassword + " " + token + " " +
				terraformConfig.AWSConfig.ClusterCIDR + " " + terraformConfig.AWSConfig.ServiceCIDR
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
