package aws

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
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
	routeRecordBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.Route53Record, aws.Route53Record})
	routeRecordBlockBody := routeRecordBlock.Body()

	zoneIDExpression := general.Data + "." + aws.Route53Zone + "." + selected + "." + zoneID
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(zoneIDExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(zoneID, values)
	routeRecordBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.ResourcePrefix))
	routeRecordBlockBody.SetAttributeValue(general.Type, cty.StringVal(CNAME))
	routeRecordBlockBody.SetAttributeValue(ttl, cty.NumberIntVal(300))

	loadBalancerExpression := "[" + aws.LoadBalancer + "." + aws.LoadBalancer + "." + dnsName + "]"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(loadBalancerExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(records, values)

	rootBody.AppendNewline()

	zoneBlock := rootBody.AppendNewBlock(general.Data, []string{aws.Route53Zone, selected})
	zoneBlockBody := zoneBlock.Body()

	zoneBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.AWSConfig.AWSRoute53Zone))
	zoneBlockBody.SetAttributeValue(privateZone, cty.BoolVal(false))
}

// CreateRoute53InternalRecord is a function that will set the AWS Route 53 record configuration in the main.tf file.
func CreateRoute53InternalRecord(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig) {
	zoneBlock := rootBody.AppendNewBlock(general.Data, []string{aws.Route53Zone, selected})
	zoneBlockBody := zoneBlock.Body()

	zoneBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.AWSConfig.AWSRoute53Zone))
	zoneBlockBody.SetAttributeValue(privateZone, cty.BoolVal(true))

	rootBody.AppendNewline()

	routeRecordBlock := rootBody.AppendNewBlock(general.Resource, []string{aws.Route53Record, aws.Route53InternalRecord})
	routeRecordBlockBody := routeRecordBlock.Body()

	zoneIDExpression := general.Data + "." + aws.Route53Zone + "." + selected + "." + zoneID
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(zoneIDExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(zoneID, values)
	routeRecordBlockBody.SetAttributeValue(name, cty.StringVal(terraformConfig.ResourcePrefix+"-internal"))
	routeRecordBlockBody.SetAttributeValue(general.Type, cty.StringVal(CNAME))
	routeRecordBlockBody.SetAttributeValue(ttl, cty.NumberIntVal(300))

	loadBalancerExpression := "[" + aws.LoadBalancer + "." + aws.InternalLoadBalancer + "." + dnsName + "]"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(loadBalancerExpression)},
	}

	routeRecordBlockBody.SetAttributeRaw(records, values)
}
