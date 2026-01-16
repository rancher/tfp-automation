package rancher2

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

const (
	audit          = "audit"
	auditVersion   = "audit_version"
	baseline       = "baseline"
	description    = "description"
	enforce        = "enforce"
	enforceVersion = "enforce_version"
	exemptions     = "exemptions"
	latest         = "latest"
	namespace      = "namespaces"
	warn           = "warn"
	warnVersion    = "warn_version"

	baselineDescription = "This is a custom baseline Pod Security Admission Configuration Template." +
		"It defines a minimally restrictive policy which prevents known privilege escalations. " +
		"This policy contains namespace level exemptions for Rancher components."
)

// SetCustomPSACT is a function that will set the Custom PSACT configurations in the main.tf file.
func SetBaselinePSACT(newFile *hclwrite.File, rootBody *hclwrite.Body, clusterName string) (*hclwrite.Body, error) {
	exemptionsNamespaces := []string{
		"ingress-nginx",
		"kube-system",
		"cattle-system",
		"cattle-epinio-system",
		"cattle-fleet-system",
		"longhorn-system",
		"cattle-neuvector-system",
		"cattle-monitoring-system",
		"rancher-alerting-drivers",
		"cis-operator-system",
		"cattle-csp-adapter-system",
		"cattle-externalip-system",
		"cattle-gatekeeper-system",
		"istio-system",
		"cattle-istio-system",
		"cattle-logging-system",
		"cattle-windows-gmsa-system",
		"cattle-sriov-system",
		"cattle-ui-plugin-system",
		"tigera-operator",
	}

	psactBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.PodSecurityAdmission, clusterName})
	psactBlockBody := psactBlock.Body()

	psactBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(clusters.RancherBaseline))
	psactBlockBody.SetAttributeValue(description, cty.StringVal(baselineDescription))

	defaultsBlock := psactBlockBody.AppendNewBlock(general.Defaults, nil)
	defaultsBlockBody := defaultsBlock.Body()

	defaultsBlockBody.SetAttributeValue(audit, cty.StringVal(baseline))
	defaultsBlockBody.SetAttributeValue(auditVersion, cty.StringVal(latest))
	defaultsBlockBody.SetAttributeValue(enforce, cty.StringVal(baseline))
	defaultsBlockBody.SetAttributeValue(enforceVersion, cty.StringVal(latest))
	defaultsBlockBody.SetAttributeValue(warn, cty.StringVal(baseline))
	defaultsBlockBody.SetAttributeValue(warnVersion, cty.StringVal(latest))

	exemptionsBlock := psactBlockBody.AppendNewBlock(exemptions, nil)
	exemptionsBlockBody := exemptionsBlock.Body()

	namespacesStr := "\"" + strings.Join(exemptionsNamespaces, "\", \"") + "\""
	namespaces := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + namespacesStr + "]")},
	}

	exemptionsBlockBody.SetAttributeRaw(namespace, namespaces)

	return rootBody, nil
}
