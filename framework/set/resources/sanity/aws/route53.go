package aws

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	CNAME       = "CNAME"
	dnsName     = "dns_name"
	privateZone = "private_zone"
	records     = "records"
	selected    = "selected"
	ttl         = "ttl"
	zoneID      = "zone_id"
)

// CreateRoute53Record is a function that will set the AWS Route 53 record configuration in the main.tf file.
func CreateRoute53Record(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	routeRecordBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Route53Record, defaults.Route53Record})
	routeRecordBlockBody := routeRecordBlock.Body()

	zoneIDExpression := defaults.Data + "." + defaults.Route53Zone + "." + selected + "." + zoneID
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(zoneIDExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(zoneID, values)
	routeRecordBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.HostnamePrefix))
	routeRecordBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(CNAME))
	routeRecordBlockBody.SetAttributeValue(ttl, cty.NumberIntVal(300))

	loadBalancerExpression := "[" + defaults.LoadBalancer + "." + defaults.LoadBalancer + "." + dnsName + "]"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(loadBalancerExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(records, values)

	rootBody.AppendNewline()

	zoneBlock := rootBody.AppendNewBlock(defaults.Data, []string{defaults.Route53Zone, selected})
	zoneBlockBody := zoneBlock.Body()

	zoneBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.AWSConfig.AWSRoute53Zone))
	zoneBlockBody.SetAttributeValue(privateZone, cty.BoolVal(false))
}

// CreateRoute53InternalRecord is a function that will set the AWS Route 53 record configuration in the main.tf file.
func CreateRoute53InternalRecord(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	routeRecordBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.Route53Record, defaults.Route53InternalRecord})
	routeRecordBlockBody := routeRecordBlock.Body()

	zoneIDExpression := defaults.Data + "." + defaults.Route53Zone + "." + selected + "." + zoneID
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(zoneIDExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(zoneID, values)
	routeRecordBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.HostnamePrefix+"-internal"))
	routeRecordBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(CNAME))
	routeRecordBlockBody.SetAttributeValue(ttl, cty.NumberIntVal(300))

	loadBalancerExpression := "[" + defaults.LoadBalancer + "." + defaults.InternalLoadBalancer + "." + dnsName + "]"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(loadBalancerExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(records, values)
}
