package rke2k3s

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	"github.com/zclconf/go-cty/cty"
)

const (
	port = "3228"
)

// SetProxyConfig is a function that will set the proxy configurations in the main.tf file.
func SetProxyConfig(clusterBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) error {
	agentVarsOneBlock := clusterBlockBody.AppendNewBlock(clusters.AgentEnvVars, nil)
	agentVarsOneBlockBody := agentVarsOneBlock.Body()

	agentVarsOneBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(httpProxy))
	agentVarsOneBlockBody.SetAttributeValue(general.Value, cty.StringVal("http://"+terraformConfig.Proxy.ProxyBastion+":"+port))

	agentVarsTwoBlock := clusterBlockBody.AppendNewBlock(clusters.AgentEnvVars, nil)
	agentVarsTwoBlockBody := agentVarsTwoBlock.Body()

	agentVarsTwoBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(httpsProxy))
	agentVarsTwoBlockBody.SetAttributeValue(general.Value, cty.StringVal("http://"+terraformConfig.Proxy.ProxyBastion+":"+port))
	agentVarsThreeBlock := clusterBlockBody.AppendNewBlock(clusters.AgentEnvVars, nil)
	agentVarsThreeBlockBody := agentVarsThreeBlock.Body()

	agentVarsThreeBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(noProxy))
	agentVarsThreeBlockBody.SetAttributeValue(general.Value, cty.StringVal(noProxyValue))
	return nil
}
