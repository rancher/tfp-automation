package imported

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/modules"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/rancher/tfp-automation/framework/set/resources/imported"
	"github.com/zclconf/go-cty/cty"
)

// importNodes is a function that will import the nodes to the cluster
func importNodes(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, nodeOnePublicDNS, kubeConfig, importCommand,
	clusterName string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	scriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/provisioning/imported/import-nodes.sh")

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

	if terraformConfig.Module == modules.ImportRKE1 {
		command += " " + kubeConfig
	}

	command += "'"

	// Need to first create a null resource block to copy the script to the node.
	nullResourceBlockBody, provisionerBlockBody := imported.CreateImportedNullResource(rootBody, terraformConfig, nodeOnePublicDNS, clusterName)

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(scriptContent) + "' > /tmp/import-nodes.sh"),
		cty.StringVal("chmod +x /tmp/import-nodes.sh"),
	}))

	var dependsOnServer string

	if terraformConfig.Module == modules.ImportK3s || terraformConfig.Module == modules.ImportRKE2 {
		dependsOnServer = `[` + defaults.NullResource + `.` + addServer + serverTwo + `, ` + defaults.NullResource + `.` + addServer + serverThree + `]`
	} else {
		dependsOnServer = `[` + defaults.RKECluster + `.` + defaults.RKECluster + `]`
	}

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	// A second null resource block is needed to properly run the script on the node. This is because the cluster registration
	// token and RKE1 kube config will be not passed correctly as Bash parameters.
	nullResourceBlockBody, provisionerBlockBody = imported.CreateImportedNullResource(rootBody, terraformConfig, nodeOnePublicDNS, cluster)

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte(command)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	dependsOnServer = `[` + defaults.NullResource + `.` + clusterName + `]`

	server = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnServer)},
	}

	nullResourceBlockBody.SetAttributeRaw(defaults.DependsOn, server)

	return nil
}
