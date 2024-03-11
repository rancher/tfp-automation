package provisioning

import (
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	baseline    = "baseline"
	description = "This is a custom baseline Pod Security Admission Configuration Template." + 
	              "It defines a minimally restrictive policy which prevents known privilege escalations. " +
	              "This policy contains namespace level exemptions for Rancher components."
	latest      = "latest"
)

var exemptionsNamespaces = []string{
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

// SetCustomPSACT is a function that will set the Custom PSACT configurations in the main.tf file.
func SetBaselinePSACT(newFile *hclwrite.File, rootBody *hclwrite.Body) (*hclwrite.File, *hclwrite.Body) {
	psactBlock := rootBody.AppendNewBlock("resource", []string{"rancher2_pod_security_admission_configuration_template", "rancher2_pod_security_admission_configuration_template"})
	psactBlockBody := psactBlock.Body()

	psactBlockBody.SetAttributeValue("name", cty.StringVal("rancher-baseline"))
	psactBlockBody.SetAttributeValue("description", cty.StringVal(description))

	defaultsBlock := psactBlockBody.AppendNewBlock("defaults", nil)
	defaultsBlockBody := defaultsBlock.Body()

	defaultsBlockBody.SetAttributeValue("audit", cty.StringVal(baseline))
	defaultsBlockBody.SetAttributeValue("audit_version", cty.StringVal(latest))
	defaultsBlockBody.SetAttributeValue("enforce", cty.StringVal(baseline))
	defaultsBlockBody.SetAttributeValue("enforce_version", cty.StringVal(latest))
	defaultsBlockBody.SetAttributeValue("warn", cty.StringVal(baseline))
	defaultsBlockBody.SetAttributeValue("warn_version", cty.StringVal(latest))

	exemptionsBlock := psactBlockBody.AppendNewBlock("exemptions", nil)
	exemptionsBlockBody := exemptionsBlock.Body()

	namespacesStr := "\"" + strings.Join(exemptionsNamespaces, "\", \"") + "\""
	namespaces := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("[" + namespacesStr + "]")},
	}
	
	exemptionsBlockBody.SetAttributeRaw("namespaces", namespaces)

	return newFile, rootBody
}
