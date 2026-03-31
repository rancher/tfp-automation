package nodedriver

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/clustertypes"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
)

// setRKEConfig is a function that will set the RKE configurations in the main.tf file.
func setRKEConfig(clusterBlockBody *hclwrite.Body, terraformConfig *config.TerraformConfig) (*hclwrite.Body, error) {
	rkeConfigBlock := clusterBlockBody.AppendNewBlock(clusters.RkeConfig, nil)
	rkeConfigBlockBody := rkeConfigBlock.Body()

	if terraformConfig.ChartValues != "" {
		chartValues := hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOF\n" + terraformConfig.ChartValues + "\nEOF"},
		})

		rkeConfigBlockBody.SetAttributeRaw(clusters.ChartValues, chartValues)
	}

	var cidrValues hclwrite.Tokens

	if terraformConfig.AWSConfig.EnablePrimaryIPv6 {
		if strings.Contains(terraformConfig.Module, clustertypes.K3S) {
			cidrValues = hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "<<EOF\ncluster-cidr: " + terraformConfig.AWSConfig.ClusterCIDR + "\nservice-cidr: " +
					terraformConfig.AWSConfig.ServiceCIDR + "\nflannel-ipv6-masq: true" + "\ningress-controller: \"traefik\"\nEOF"},
			})
		} else {
			cidrValues = hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "<<EOF\ncluster-cidr: " + terraformConfig.AWSConfig.ClusterCIDR + "\nservice-cidr: " +
					terraformConfig.AWSConfig.ServiceCIDR + "\ningress-controller: \"traefik\"\nEOF"},
			})
		}

	} else {
		cidrValues = hclwrite.TokensForTraversal(hcl.Traversal{
			hcl.TraverseRoot{Name: "<<EOF\ningress-controller: \"traefik\"\nEOF"},
		})
	}

	rkeConfigBlockBody.SetAttributeRaw(clusters.MachineGlobalConfig, cidrValues)

	return rkeConfigBlockBody, nil
}
