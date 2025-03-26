package aws

import (
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/framework/set/defaults"
	"github.com/zclconf/go-cty/cty"
)

const (
	loadBalancerARN = "load_balancer_arn"
	forward         = "forward"
	targetGroupARN  = "target_group_arn"
)

// CreateLoadBalancerListeners is a function that will set the load balancer listeners configurations in the main.tf file.
func CreateLoadBalancerListeners(rootBody *hclwrite.Body, port int64) {
	listenersGroupBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LoadBalancerListener, defaults.LoadBalancerListener + "_" + strconv.FormatInt(port, 10)})
	listenersGroupBlockBody := listenersGroupBlock.Body()

	loadBalancerExpression := defaults.LoadBalancer + "." + defaults.LoadBalancer + ".arn"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(loadBalancerExpression)},
	}

	listenersGroupBlockBody.SetAttributeRaw(loadBalancerARN, values)
	listenersGroupBlockBody.SetAttributeValue(defaults.Port, cty.NumberIntVal(port))
	listenersGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(TCP))

	defaultActionBlock := listenersGroupBlockBody.AppendNewBlock(defaults.DefaultAction, nil)
	defaultActionBlockBody := defaultActionBlock.Body()

	defaultActionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(forward))

	targetGroupExpression := defaults.LoadBalancerTargetGroup + "." + defaults.TargetGroupPrefix + strconv.FormatInt(port, 10) + ".arn"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetGroupExpression)},
	}

	defaultActionBlockBody.SetAttributeRaw(targetGroupARN, values)
}

// CreateInternalLoadBalancerListeners is a function that will set the internal load balancer listeners configurations in the main.tf file.
func CreateInternalLoadBalancerListeners(rootBody *hclwrite.Body, port int64) {
	listenersGroupBlock := rootBody.AppendNewBlock(defaults.Resource, []string{defaults.LoadBalancerListener, defaults.LoadBalancerInternalListerner + "_" + strconv.FormatInt(port, 10)})
	listenersGroupBlockBody := listenersGroupBlock.Body()

	loadBalancerExpression := defaults.LoadBalancer + "." + defaults.InternalLoadBalancer + ".arn"
	values := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(loadBalancerExpression)},
	}

	listenersGroupBlockBody.SetAttributeRaw(loadBalancerARN, values)
	listenersGroupBlockBody.SetAttributeValue(defaults.Port, cty.NumberIntVal(port))
	listenersGroupBlockBody.SetAttributeValue(protocol, cty.StringVal(TCP))

	defaultActionBlock := listenersGroupBlockBody.AppendNewBlock(defaults.DefaultAction, nil)
	defaultActionBlockBody := defaultActionBlock.Body()

	defaultActionBlockBody.SetAttributeValue(defaults.Type, cty.StringVal(forward))

	targetGroupExpression := defaults.LoadBalancerTargetGroup + "." + defaults.TargetGroupInternalPrefix + strconv.FormatInt(port, 10) + ".arn"
	values = hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(targetGroupExpression)},
	}

	defaultActionBlockBody.SetAttributeRaw(targetGroupARN, values)
}
