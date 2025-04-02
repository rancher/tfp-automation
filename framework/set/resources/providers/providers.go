package providers

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/linode"
	"github.com/sirupsen/logrus"
)

type ProviderResourceFunc func(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error)

type ProviderResources struct {
	CreateAirgap    ProviderResourceFunc
	CreateNonAirgap ProviderResourceFunc
}

// TunnelToProvider returns an struct that allows a user to create resources from a given provider
func TunnelToProvider(terraformConfig *config.TerraformConfig) ProviderResources {
	switch terraformConfig.NodeProvider {
	case providers.AWS:
		logrus.Infof("Creating AWS resources...")
		return ProviderResources{
			CreateAirgap:    aws.CreateAirgappedAWSResources,
			CreateNonAirgap: aws.CreateAWSResources,
		}
	case providers.Linode:
		logrus.Infof("Creating Linode resources...")
		return ProviderResources{
			CreateNonAirgap: linode.CreateLinodeResources,
		}
	default:
		panic(fmt.Sprintf("Unsupported provider: %s", terraformConfig.NodeProvider))
	}

}
