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
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/provisioning/imported/nullresource"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// importNodes is a function that will import the nodes to the cluster
func importNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, nodeOnePublicDNS, kubeConfig, importCommand string) error {
	userDir, _ := rancher2.SetKeyPath(keypath.AirgapKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

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

	kubeConfig = `\"` + kubeConfig + `\"`
	importCommand = `\"` + importCommand + `\"`

	command := "bash -c '/tmp/import-nodes.sh " + encodedPEMFile + " " + terraformConfig.Standalone.OSUser + " " +
		terraformConfig.Standalone.OSGroup + " " + nodeOnePublicDNS + " " + importCommand

	if strings.Contains(terraformConfig.Module, clustertypes.RKE1) && strings.Contains(terraformConfig.Module, defaults.Import) {
		command += " " + kubeConfig
	}

	command += "'"

	// Need to first create a null resource block to copy the script to the node.
	copyScriptName := terraformConfig.ResourcePrefix + `_` + copyScript
	nullResourceBlockBody, provisionerBlockBody := nullresource.CreateImportedNullResource(rootBody, terraformConfig, nodeOnePublicDNS, copyScriptName)

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(scriptContent) + "' > /tmp/import-nodes.sh"),
		cty.StringVal("chmod +x /tmp/import-nodes.sh"),
	}))

	var dependsOnServer string

	if strings.Contains(terraformConfig.Module, defaults.Import) && !strings.Contains(terraformConfig.Module, clustertypes.RKE1) &&
		!strings.Contains(terraformConfig.Module, clustertypes.WINDOWS) {
		addServerTwoName := addServer + terraformConfig.ResourcePrefix + `_` + serverTwo
		addServerThreeName := addServer + terraformConfig.ResourcePrefix + `_` + serverThree
		dependsOnServer = `[` + defaults.NullResource + `.` + addServerTwoName + `, ` + defaults.NullResource + `.` + addServerThreeName + `]`
	} else if terraformConfig.Module == modules.ImportEC2RKE2Windows {
		dependsOnServer = `[` + defaults.TimeSleep + `.` + defaults.TimeSleep + `-` + terraformConfig.ResourcePrefix + `]`
	} else {
		dependsOnServer = `[` + defaults.RKECluster + `.` + terraformConfig.ResourcePrefix + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	// A second null resource block is needed to properly run the script on the node. This is because the cluster registration
	// token and RKE1 kube config will be not passed correctly as Bash parameters.
	importClusterName := terraformConfig.ResourcePrefix + `_` + importCluster
	nullResourceBlockBody, provisionerBlockBody = nullresource.CreateImportedNullResource(rootBody, terraformConfig, nodeOnePublicDNS, importClusterName)

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

	return nil
}
