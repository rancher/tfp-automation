package providers

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rancher/tfp-automation/config"
	"github.com/rancher/tfp-automation/defaults/providers"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/aws"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/harvester"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/linode"
	"github.com/rancher/tfp-automation/framework/set/resources/providers/vsphere"
	"github.com/sirupsen/logrus"
)

type ProviderResourceFunc func(file *os.File, newFile *hclwrite.File, tfBlockBody, rootBody *hclwrite.Body, terraformConfig *config.TerraformConfig,
	terratestConfig *config.TerratestConfig, instances []string) (*os.File, error)

type ProviderResources struct {
	CreateAirgap    ProviderResourceFunc
	CreateNonAirgap ProviderResourceFunc
	CreateIPv6      ProviderResourceFunc
}

// TunnelToProvider returns an struct that allows a user to create resources from a given provider
func TunnelToProvider(provider string) ProviderResources {
	switch provider {
	case providers.AWS:
		logrus.Infof("Creating AWS resources...")
		return ProviderResources{
			CreateAirgap:    aws.CreateAirgappedAWSResources,
			CreateNonAirgap: aws.CreateAWSResources,
			CreateIPv6:      aws.CreateIPv6AWSResources,
		}
	case providers.Linode:
		logrus.Infof("Creating Linode resources...")
		return ProviderResources{
			CreateNonAirgap: linode.CreateLinodeResources,
		}
	case providers.Harvester:
		logrus.Infof("Creating Harvester resources...")
		return ProviderResources{
			CreateNonAirgap: harvester.CreateHarvesterResources,
		}
	case providers.Vsphere:
		logrus.Infof("Creating vSphere resources...")
		return ProviderResources{
			CreateNonAirgap: vsphere.CreateVsphereResources,
		}
	default:
		panic(fmt.Sprintf("Unsupported provider: %s", provider))
	}

}
