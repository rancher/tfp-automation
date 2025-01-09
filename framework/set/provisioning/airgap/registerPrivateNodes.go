package airgap

import (
	"encoding/base64"
	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
)

// registerPrivateNodes is a function that will register the private nodes to the cluster
func registerPrivateNodes(provisionerBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, bastionPublicIP, nodePrivateIP,
	registrationCommand string) error {
	privateKey, err := os.ReadFile(terraformConfig.PrivateKeyPath)
	if err != nil {
		return nil
	}

	encodedPEMFile := base64.StdEncoding.EncodeToString([]byte(privateKey))

	newCommand := `\"` + registrationCommand + `\"`

	provisionerBlockBody.SetAttributeRaw(defaults.Inline, hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`["`), SpacesBefore: 1},
		{Type: hclsyntax.TokenStringLit, Bytes: []byte("/tmp/register-nodes.sh " + encodedPEMFile + " " +
			terraformConfig.Standalone.RKE2User + " " + terraformConfig.Standalone.RKE2Group + " " + bastionPublicIP + " " +
			nodePrivateIP + " " + newCommand)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"]`), SpacesBefore: 1},
	})

	return nil
}
