package aws

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/providers/aws"
	"github.com/zclconf/go-cty/cty"
)

const (
	eachValue   = "each.value"
	forEach     = "for_each"
	instanceIDs = "instance_ids"
)

// CreateTargetGroupAttachments is a function that will set the target group attachments configurations in the main.tf file.
func CreateTargetGroupAttachments(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, lbTargetGroupAttachment,
	targetGroupAttachmentServer string, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(general.Resource, []string{lbTargetGroupAttachment, targetGroupAttachmentServer})
	targetGroupBlockBody := targetGroupBlock.Body()

	instanceValueExpression := general.Local + "." + instanceIDs
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(instanceValueExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(forEach, values)

	targetGroupExpression := aws.LoadBalancerTargetGroup + "." + aws.TargetGroupPrefix + fmt.Sprint(port) + ".arn"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetGroupExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(aws.TargetGroupARN, values)

	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(eachValue)},
	}

	targetGroupBlockBody.SetAttributeRaw(aws.TargetID, values)
	targetGroupBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
}

// CreateInternalTargetGroupAttachments is a function that will set the internal target group attachments configurations in the main.tf file.
func CreateInternalTargetGroupAttachments(rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig, lbTargetGroupAttachment,
	targetGroupAttachmentServer string, port int64) {
	targetGroupBlock := rootBody.AppendNewBlock(general.Resource, []string{lbTargetGroupAttachment, targetGroupAttachmentServer})
	targetGroupBlockBody := targetGroupBlock.Body()

	instanceValueExpression := general.Local + "." + instanceIDs
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(instanceValueExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(forEach, values)

	targetGroupExpression := aws.LoadBalancerTargetGroup + "." + aws.TargetGroupInternalPrefix + fmt.Sprint(port) + ".arn"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetGroupExpression)},
	}

	targetGroupBlockBody.SetAttributeRaw(aws.TargetGroupARN, values)

	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(eachValue)},
	}

	targetGroupBlockBody.SetAttributeRaw(aws.TargetID, values)
	targetGroupBlockBody.SetAttributeValue(general.Port, cty.NumberIntVal(port))
}
