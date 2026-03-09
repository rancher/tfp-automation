package hosted

import (
	"os"
	"strconv"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/resourceblocks/nodeproviders/google"
	"github.com/rancher/tfp-automation/framework/set/defaults/general"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2"
	"github.com/rancher/tfp-automation/framework/set/defaults/rancher2/clusters"
	resources "github.com/rancher/tfp-automation/framework/set/resources/rancher2"
	"github.com/zclconf/go-cty/cty"
)

const (
	active                   = "active"
	autoRepair               = "auto_repair"
	autoUpgrade              = "auto_upgrade"
	builtin                  = "builtin"
	googleDriver             = "google"
	httpLoadBalancing        = "http_load_balancing"
	horizontalPodAutoscaling = "horizontal_pod_autoscaling"
	url                      = "url"
)

// SetGKE is a function that will set the GKE configurations in the main.tf file.
func SetGKE(terraformConfig *config.TerraformConfig, terratestConfig *config.TerratestConfig, newFile *hclwrite.File, rootBody *hclwrite.Body,
	file *os.File) (*hclwrite.File, *os.File, error) {
	nodeDriverBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.NodeDriver, rancher2.NodeDriver})
	nodeDriverBlockBody := nodeDriverBlock.Body()

	provider := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(general.Rancher2 + "." + general.AdminUser)},
	}

	nodeDriverBlockBody.SetAttributeRaw(general.Provider, provider)
	nodeDriverBlockBody.SetAttributeValue(active, cty.BoolVal(true))
	nodeDriverBlockBody.SetAttributeValue(builtin, cty.BoolVal(true))
	nodeDriverBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(googleDriver))
	nodeDriverBlockBody.SetAttributeValue(url, cty.StringVal("local://"))

	rootBody.AppendNewline()

	cloudCredBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.CloudCredential, rancher2.CloudCredential})
	cloudCredBlockBody := cloudCredBlock.Body()

	cloudCredBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	googleCredConfigBlock := cloudCredBlockBody.AppendNewBlock(google.GoogleCredentialConfig, nil)
	googleCredConfigBlockBody := googleCredConfigBlock.Body()

	googleCredConfigBlockBody.SetAttributeValue(google.AuthEncodedJSON, cty.StringVal(terraformConfig.GoogleCredentials.AuthEncodedJSON))

	dependsOnBlock := `[` + rancher2.NodeDriver + `.` + rancher2.NodeDriver + `]`

	server := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(dependsOnBlock)},
	}

	cloudCredBlockBody.SetAttributeRaw(general.DependsOn, server)

	rootBody.AppendNewline()

	clusterBlock := rootBody.AppendNewBlock(general.Resource, []string{rancher2.Cluster, rancher2.Cluster})
	clusterBlockBody := clusterBlock.Body()

	clusterBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	gkeConfigBlock := clusterBlockBody.AppendNewBlock(google.GKEConfig, nil)
	gkeConfigBlockBody := gkeConfigBlock.Body()

	gkeConfigBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix))

	cloudCredSecret := hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(rancher2.CloudCredential + "." + rancher2.CloudCredential + ".id")},
	}

	gkeConfigBlockBody.SetAttributeRaw(google.GoogleCredentialSecret, cloudCredSecret)
	gkeConfigBlockBody.SetAttributeValue(google.Region, cty.StringVal(terraformConfig.GoogleConfig.Region))
	gkeConfigBlockBody.SetAttributeValue(google.ProjectID, cty.StringVal(terraformConfig.GoogleConfig.ProjectID))
	gkeConfigBlockBody.SetAttributeValue(clusters.KubernetesVersion, cty.StringVal(terratestConfig.KubernetesVersion))
	gkeConfigBlockBody.SetAttributeValue(google.Network, cty.StringVal(terraformConfig.GoogleConfig.Network))
	gkeConfigBlockBody.SetAttributeValue(google.Subnetwork, cty.StringVal(terraformConfig.GoogleConfig.Subnetwork))

	clusterAddOnsBlock := gkeConfigBlockBody.AppendNewBlock(google.ClusterAddOns, nil)
	clusterAddOnsBlockBody := clusterAddOnsBlock.Body()

	clusterAddOnsBlockBody.SetAttributeValue(httpLoadBalancing, cty.BoolVal(true))
	clusterAddOnsBlockBody.SetAttributeValue(horizontalPodAutoscaling, cty.BoolVal(true))

	for count, pool := range terratestConfig.Nodepools {
		poolNum := strconv.Itoa(count)

		_, err := resources.SetResourceNodepoolValidation(terraformConfig, pool, poolNum)
		if err != nil {
			return nil, nil, err
		}

		nodePoolsBlock := gkeConfigBlockBody.AppendNewBlock(google.NodePools, nil)
		nodePoolsBlockBody := nodePoolsBlock.Body()

		nodePoolsBlockBody.SetAttributeValue(google.InitialNodeCount, cty.NumberIntVal(pool.Quantity))
		nodePoolsBlockBody.SetAttributeValue(google.MaxPodsConstraint, cty.NumberIntVal(pool.MaxPodsConstraint))
		nodePoolsBlockBody.SetAttributeValue(general.ResourceName, cty.StringVal(terraformConfig.ResourcePrefix+`-pool`+poolNum))
		nodePoolsBlockBody.SetAttributeValue(google.Version, cty.StringVal(terratestConfig.KubernetesVersion))

		configBlock := nodePoolsBlockBody.AppendNewBlock(google.Config, nil)
		configBlockBody := configBlock.Body()

		configBlockBody.SetAttributeValue(google.ImageType, cty.StringVal(terraformConfig.GoogleConfig.ImageType))
		configBlockBody.SetAttributeValue(google.MachineType, cty.StringVal(terraformConfig.GoogleConfig.MachineType))
		configBlockBody.SetAttributeValue(google.DiskSizeGb, cty.NumberIntVal(terraformConfig.GoogleConfig.Size))

		managementBlock := nodePoolsBlockBody.AppendNewBlock(google.Management, nil)
		managementBlockBody := managementBlock.Body()

		managementBlockBody.SetAttributeValue(autoRepair, cty.BoolVal(true))
		managementBlockBody.SetAttributeValue(autoUpgrade, cty.BoolVal(true))
	}

	return newFile, file, nil
}
