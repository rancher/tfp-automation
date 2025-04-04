package harvester

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	locals            = "locals"
	requiredProviders = "required_providers"
	k3sServerOne      = "k3s_server1"
	k3sServerTwo      = "k3s_server2"
	k3sServerThree    = "k3s_server3"
	rke2ServerOne     = "rke2_server1"
	rke2ServerTwo     = "rke2_server2"
	rke2ServerThree   = "rke2_server3"
)

// CreateTerraformProviderBlock will up the terraform block with the required harvester provider.
func CreateTerraformProviderBlock(tfBlockBody *hclwrite.Body) {
	harvesterProviderVersion := os.Getenv("HARVESTER_PROVIDER_VERSION")
	kubernetesProviderVersion := os.Getenv("KUBERNETES_PROVIDER_VERSION")

	reqProvsBlock := tfBlockBody.AppendNewBlock(requiredProviders, nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue("harvester", cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.HarvesterSource),
		defaults.Version: cty.StringVal(harvesterProviderVersion),
	}))

	reqProvsBlockBody.SetAttributeValue("kubernetes", cty.ObjectVal(map[string]cty.Value{
		defaults.Source:  cty.StringVal(defaults.KubernetesSource),
		defaults.Version: cty.StringVal(kubernetesProviderVersion),
	}))
}

// CreateHarvesterProviderBlock will set up the harvester provider block.
func CreateHarvesterProviderBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	harvesterProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{terraformConfig.Provider})
	harvesterProvBlockBody := harvesterProvBlock.Body()

	pathModuleVar := fmt.Sprint(`"${local.codebase_root_path}/local.yaml"`)
	hclPathModule := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(pathModuleVar)},
	}

	harvesterProvBlockBody.SetAttributeRaw("kubeconfig", hclPathModule)

	kubernetesProvBlock := rootBody.AppendNewBlock(defaults.Provider, []string{defaults.Kubernetes})
	kubernetesProvBlockBody := kubernetesProvBlock.Body()
	kubernetesProvBlockBody.SetAttributeRaw("config_path", hclPathModule)

}

// CreateLocalBlock will set up the local block. Returns the local block.
func CreateLocalBlock(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	localBlock := rootBody.AppendNewBlock(locals, nil)
	localBlockBody := localBlock.Body()

	pathModuleVar := fmt.Sprint(`abspath("${path.module}")`)
	hclPathModule := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(pathModuleVar)},
	}
	localBlockBody.SetAttributeRaw(defaults.CodebaseRootPath, hclPathModule)

	pathModule := fmt.Sprint(`abspath(path.module)`)
	hclPath2Module := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(pathModule)},
	}
	localBlockBody.SetAttributeRaw(defaults.ModulePath, hclPath2Module)

	relPath := fmt.Sprint(`substr(local.module_path, length(local.codebase_root_path)+1, length(local.module_path))`)
	relPathModule := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(relPath)},
	}
	localBlockBody.SetAttributeRaw(defaults.ModuleRelPath, relPathModule)

	publicKey := getPublicSSHKey(terraformConfig.PrivateKeyPath)
	localBlockBody.SetAttributeRaw(defaults.CloudInit, hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{
			Name: fmt.Sprintf("<<-EOT\n#cloud-config\npackage_update: true\npackages:\n  - qemu-guest-agent\nruncmd:\n  - - systemctl\n    - enable\n    - --now\n    - qemu-guest-agent.service\nssh_authorized_keys:\n  - %s\nEOT", publicKey),
		},
	}))
}
