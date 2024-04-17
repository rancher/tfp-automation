package provisioning

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	blocks "github.com/rancher/tfp-automation/defaults/resourceblocks"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/psact"
	"github.com/zclconf/go-cty/cty"
)

// SetCustomPSACT is a function that will set the Custom PSACT configurations in the main.tf file.
func SetBaselinePSACT(newFile *hclwrite.File, rootBody *hclwrite.Body) (*hclwrite.File, *hclwrite.Body) {
	psactBlock := rootBody.AppendNewBlock(blocks.ResourceName, []string{blocks.PodSecurityAdmission, blocks.PodSecurityAdmission})
	psactBlockBody := psactBlock.Body()

	psactBlockBody.SetAttributeValue(blocks.ResourceName, cty.StringVal(psact.RancherBaseline))
	psactBlockBody.SetAttributeValue(psact.Description, cty.StringVal(psact.BaselineDescription))

	defaultsBlock := psactBlockBody.AppendNewBlock(psact.Defaults, nil)
	defaultsBlockBody := defaultsBlock.Body()

	defaultsBlockBody.SetAttributeValue(psact.Audit, cty.StringVal(psact.Baseline))
	defaultsBlockBody.SetAttributeValue(psact.AuditVersion, cty.StringVal(psact.Latest))
	defaultsBlockBody.SetAttributeValue(psact.Enforce, cty.StringVal(psact.Baseline))
	defaultsBlockBody.SetAttributeValue(psact.EnforceVersion, cty.StringVal(psact.Latest))
	defaultsBlockBody.SetAttributeValue(psact.Warn, cty.StringVal(psact.Baseline))
	defaultsBlockBody.SetAttributeValue(psact.WarnVersion, cty.StringVal(psact.Latest))

	exemptionsBlock := psactBlockBody.AppendNewBlock(psact.Exemptions, nil)
	exemptionsBlockBody := exemptionsBlock.Body()

	namespacesStr := "\"" + strings.Join(psact.ExemptionsNamespaces, "\", \"") + "\""
	namespaces := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + namespacesStr + "]")},
	}

	exemptionsBlockBody.SetAttributeRaw(psact.Namespaces, namespaces)

	return newFile, rootBody
}
