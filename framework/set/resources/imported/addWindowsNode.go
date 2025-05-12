package imported

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported/nullresource"
	"github.com/zclconf/go-cty/cty"
)

// AddWindowsNodeToImportedCluster is a helper function that will add an additional Windows node to the initial server.
func AddWindowsNodeToImportedCluster(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	serverOnePrivateIP,
	windowsNodePublicDNS, token string) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	serverScriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/resources/rke2/add-wins.ps1")

	serverOneScriptContent, err := os.ReadFile(serverScriptPath)
	if err != nil {
		return
	}

	addImportedWindowsNode(rootBody, terraformConfig, terratestConfig, serverOnePrivateIP, windowsNodePublicDNS, token, serverOneScriptContent)
}

// addImportedWindowsNode is a helper function that will add an additional Windows node to the initial server.
func addImportedWindowsNode(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig,
	serverOnePrivateIP, windowsNodePublicDNS,
	token string, script []byte) {
	copyScriptName := terraformConfig.ResourcePrefix + copyScript + windowsServer

	nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedWindowsNullResource(rootBody, terraformConfig, terratestConfig, windowsNodePublicDNS, copyScriptName)
	rootBody.AppendNewline()

	dependsOnServer := `[` + defaults.AwsInstance + `.` + terraformConfig.ResourcePrefix + `-windows` + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	// Due to nuances in Powershell with copying the script over, we need to split the script into lines and echo each line to the file.
	// This is a workaround for the issue where the script is not being copied correctly as the Bash scripts typically are copied over.
	var inlineCommands []cty.Value
	scriptLines := strings.Split(string(script), "\n")

	for i, scriptLine := range scriptLines {
		if strings.TrimSpace(scriptLine) == "" {
			continue
		}

		command := "echo " + scriptLine

		if i == 0 {
			command += " > C:\\Windows\\Temp\\init-server.ps1"
		} else {
			command += " >> C:\\Windows\\Temp\\init-server.ps1"
		}

		inlineCommands = append(inlineCommands, cty.StringVal(command))
	}

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal(inlineCommands))
	nullResourceBlockBody, provisionerBlockBody = nullresource.CreateImportedWindowsNullResource(rootBody, terraformConfig, terratestConfig, windowsNodePublicDNS, addWindowsNode)
	version := terraformConfig.Standalone.RKE2Version

	command := "powershell.exe -File C:\\\\Windows\\\\Temp\\\\init-server.ps1 -ArgumentList -K8S_VERSION " + version + " -RKE2_SERVER_IP " + serverOnePrivateIP +
		" -RKE2_TOKEN " + token

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	dependsOnServer = `[` + defaults.NullResource + `.` + copyScriptName + `]`

	server = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)
}
