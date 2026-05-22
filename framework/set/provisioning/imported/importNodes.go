package imported

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

const (
	addServer     = "add_server_"
	addAgent      = "add_agent_"
	createCluster = "create_cluster"
	importCluster = "import_cluster"
	copyScript    = "copy_script"
)

// ImportNodes is a function that will import the nodes to the cluster
func ImportNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	nodeOnePublicIP, importCommand string, additionalServerNodeNames, additionalAgentNodeNames []string) error {
	userDir, _ := rancher2.SetKeyPath(keypath.RancherKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	scriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/provisioning/imported/import-nodes.sh")

	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	privateKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	if err != nil {
		return err
	}

	encodedPEMFile := base64.StdEncoding.EncodeToString([]byte(privateKey))

	importCommand = `\"` + importCommand + `\"`

	command := "bash -c '/tmp/import-nodes.sh " + encodedPEMFile + " " + terraformConfig.Standalone.OSUser + " " +
		terraformConfig.Standalone.OSGroup + " " + importCommand + "'"

	linuxNodeNames := make([]string, 0, len(additionalServerNodeNames)+len(additionalAgentNodeNames)+1)
	linuxNodeNames = append(linuxNodeNames, terraformConfig.ResourcePrefix+"_server1")
	linuxNodeNames = append(linuxNodeNames, additionalServerNodeNames...)
	linuxNodeNames = append(linuxNodeNames, additionalAgentNodeNames...)

	// Need to first create a null resource block to copy the script to the node.
	copyScriptName := terraformConfig.ResourcePrefix + `_` + copyScript
	nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, nodeOnePublicIP, copyScriptName, linuxNodeNames)

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(scriptContent) + "' > /tmp/import-nodes.sh"),
		cty.StringVal("chmod +x /tmp/import-nodes.sh"),
	}))

	var dependsOnServer string

	switch terraformConfig.Module {
	case modules.ImportedAWSRKE2Windows2019, modules.ImportedAWSRKE2Windows2022:
		dependsOnServer = `[` + general.TimeSleep + `.` + general.TimeSleep + `-` + terraformConfig.ResourcePrefix + `-import_wins` + `]`
	default:
		if len(additionalServerNodeNames) == 0 && len(additionalAgentNodeNames) == 0 {
			dependsOnServer = fmt.Sprintf("[%s.%s_%s]", general.NullResource, terraformConfig.ResourcePrefix, createCluster)
		} else {
			dependsOnResources := make([]string, 0, len(additionalServerNodeNames)+len(additionalAgentNodeNames))
			for _, nodeName := range additionalServerNodeNames {
				dependsOnResources = append(dependsOnResources, fmt.Sprintf("%s.%s%s", general.NullResource, addServer, nodeName))
			}

			for _, nodeName := range additionalAgentNodeNames {
				dependsOnResources = append(dependsOnResources, fmt.Sprintf("%s.%s%s", general.NullResource, addAgent, nodeName))
			}

			dependsOnServer = fmt.Sprintf("[%s]", strings.Join(dependsOnResources, ", "))
		}
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

	// A second null resource block is needed to properly run the script on the node. This is because the cluster registration
	// token will not be passed correctly as Bash parameters.
	importClusterName := terraformConfig.ResourcePrefix + `_` + importCluster
	nullResourceBlockBody, provisionerBlockBody = nullresource.CreateImportedNullResource(rootBody, terraformConfig, nodeOnePublicIP, importClusterName, linuxNodeNames)

	provisionerBlockBody.SetAttributeRaw(general.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	dependsOnServer = `[` + general.NullResource + `.` + copyScriptName + `]`

	server = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(general.DependsOn, server)

	return nil
}
