package airgap

import (
	"encoding/base64"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
)

// RegisterPrivateNodes is a function that will register the private nodes to the cluster
func RegisterPrivateNodes(provisionerBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, nodePrivateIP,
	registrationCommand string) error {
	privateKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	if err != nil {
		return nil
	}

	encodedPEMFile := base64.StdEncoding.EncodeToString([]byte(privateKey))

	newCommand := `\"` + registrationCommand + `\"`

	provisionerBlockBody.SetAttributeRaw(general.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte("/tmp/register-nodes.sh " + encodedPEMFile + " " +
			terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " +
			nodePrivateIP + " " + newCommand + " " + terraformConfig.PrivateRegistries.SystemDefaultRegistry + " || true")},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	return nil
}

// RegisterWindowsPrivateNodes is a function that will register the private  Windows nodes to the cluster
func RegisterWindowsPrivateNodes(provisionerBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, nodePrivateIP,
	registrationCommand string) error {
	windowsPrivateKey, err := os.ReadFile(terraformConfig.WindowsPrivateKeyPath)
	if err != nil {
		return nil
	}

	encodedWindowsPEMFile := base64.StdEncoding.EncodeToString([]byte(windowsPrivateKey))

	newCommand := `\"` + registrationCommand + `\"`

	provisionerBlockBody.SetAttributeRaw(general.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte("/tmp/register-windows-nodes.sh " + encodedWindowsPEMFile + " " +
			terraformConfig.Standalone.OSUser + " " + terraformConfig.Standalone.OSGroup + " " + terraformConfig.AWSConfig.WindowsAWSUser + " " +
			nodePrivateIP + " " + newCommand + " " + terraformConfig.PrivateRegistries.SystemDefaultRegistry)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	return nil
}
