package linode

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	cname      = "CNAME"
	domain     = "domain"
	domainID   = "domain_id"
	master     = "master"
	recordType = "record_type"
	soaEmail   = "soa_email"
	target     = "target"
)

// CreateDomainRecord is a function that will set the Linode domain and domain record configuration in the main.tf file.
func CreateDomainRecord(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	domainBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LinodeDomain, defaults.LinodeDomain})
	domainBlockBody := domainBlock.Body()

	domainBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(master))
	domainBlockBody.SetAttributeValue(domain, cty.StringVal(terraformConfig.LinodeConfig.Domain))
	domainBlockBody.SetAttributeValue(soaEmail, cty.StringVal(terraformConfig.LinodeConfig.SOAEmail))

	rootBody.AppendNewline()

	domainRecordBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LinodeDomainRecord, defaults.LinodeDomainRecord})
	domainRecordBlockBody := domainRecordBlock.Body()

	expression := defaults.LinodeDomain + `.` + defaults.LinodeDomain + `.id`
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(expression)},
	}

	domainRecordBlockBody.SetAttributeRaw(domainID, values)
	domainRecordBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.LinodeConfig.Domain))
	domainRecordBlockBody.SetAttributeValue(recordType, cty.StringVal(cname))
	domainRecordBlockBody.SetAttributeValue(target, cty.StringVal(terraformConfig.ResourcePrefix+"."+terraformConfig.LinodeConfig.Domain))
}
