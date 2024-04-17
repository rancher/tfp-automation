package psact

var ExemptionsNamespaces = []string{
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

const (
	Audit           = "audit"
	AuditVersion    = "audit_version"
	Baseline        = "baseline"
	Defaults        = "defaults"
	Description     = "description"
	Enforce         = "enforce"
	EnforceVersion  = "enforce_version"
	Exemptions      = "exemptions"
	Latest          = "latest"
	Namespaces      = "namespaces"
	RancherBaseline = "rancher-baseline"
	Warn            = "warn"
	WarnVersion     = "warn_version"

	BaselineDescription = "This is a custom baseline Pod Security Admission Configuration Template." +
		"It defines a minimally restrictive policy which prevents known privilege escalations. " +
		"This policy contains namespace level exemptions for Rancher components."
)
