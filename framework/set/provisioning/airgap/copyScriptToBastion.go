package airgap

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/keypath"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

// CopyScript is a function that will copy the register scripts to the bastion node
func CopyScript(provisionerBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig) error {
	userDir, _ := rancher2.SetKeyPath(keypath.AirgapKeyPath, terratestConfig.PathToRepo, terraformConfig.Provider)

	nodesScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/provisioning/airgap/register-nodes.sh")
	windowsScriptPath := filepath.Join(userDir, terratestConfig.PathToRepo, "/framework/set/provisioning/airgap/register-windows-nodes.sh")

	nodesScriptContent, err := os.ReadFile(nodesScriptPath)
	if err != nil {
		return nil
	}

	windowsScriptContent, err := os.ReadFile(windowsScriptPath)
	if err != nil {
		return nil
	}

	provisionerBlockBody.SetAttributeValue(general.Inline, cty.ListVal([]cty.Value{
		cty.StringVal("echo '" + string(nodesScriptContent) + "' > /tmp/register-nodes.sh"),
		cty.StringVal("chmod +x /tmp/register-nodes.sh"),
		cty.StringVal("echo '" + string(windowsScriptContent) + "' > /tmp/register-windows-nodes.sh"),
		cty.StringVal("chmod +x /tmp/register-windows-nodes.sh"),
	}))

	return nil
}
