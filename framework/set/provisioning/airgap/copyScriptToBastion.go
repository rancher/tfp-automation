package airgap

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

// copyScript is a function that will copy the register-nodes.sh script to the bastion node
func copyScript(provisionerBlockBody *hclwrite.Body) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	nodesScriptPath := filepath.Join(userDir, "go/src/github.com/rancher/tfp-automation/framework/set/provisioning/airgap/register-nodes.sh")

	nodesScriptContent, err := os.ReadFile(nodesScriptPath)
	if err != nil {
		return nil
	}

	provisionerBlockBody.SetAttributeValue(defaults.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(nodesScriptContent) + "' > /tmp/register-nodes.sh"),
		cty.StringVal("chmod +x /tmp/register-nodes.sh"),
	}))

	return nil
}
